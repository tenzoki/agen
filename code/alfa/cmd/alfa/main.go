package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tenzoki/agen/alfa/internal/ai"
	"github.com/tenzoki/agen/alfa/internal/audio"
	alfacontext "github.com/tenzoki/agen/alfa/internal/context"
	"github.com/tenzoki/agen/alfa/internal/gox"
	"github.com/tenzoki/agen/alfa/internal/orchestrator"
	"github.com/tenzoki/agen/alfa/internal/project"
	"github.com/tenzoki/agen/alfa/internal/sandbox"
	"github.com/tenzoki/agen/alfa/internal/speech"
	"github.com/tenzoki/agen/alfa/internal/tools"
	"github.com/tenzoki/agen/atomic/vcr"
	"github.com/tenzoki/agen/atomic/vfs"
)

func main() {
	var (
		workdir       = flag.String("workdir", ".", "Working directory")
		configFile    = flag.String("config", "", "Config file path (default: config/ai-config.json)")
		mode          = flag.String("mode", "confirm", "Execution mode: confirm or allow-all")
		provider      = flag.String("provider", "", "AI provider override (anthropic or openai)")
		iterations    = flag.Int("max-iterations", 10, "Maximum AI iterations per request")
		enableVoice   = flag.Bool("voice", false, "Enable voice input/output (requires OPENAI_API_KEY and sox)")
		headless      = flag.Bool("headless", false, "Headless mode: enables voice + allow-all (autonomous voice agent)")
		useSandbox    = flag.Bool("sandbox", false, "Use Docker sandbox for command execution (requires Docker)")
		sandboxImage  = flag.String("sandbox-image", "golang:1.24-alpine", "Docker image for sandbox")
		projectName   = flag.String("project", "", "Project name (creates if doesn't exist)")
		listProjects  = flag.Bool("list-projects", false, "List all projects and exit")
		createProject = flag.String("create-project", "", "Create a new project and exit")
		deleteProject = flag.String("delete-project", "", "Delete a project (keeps backup) and exit")
		restoreProject = flag.String("restore-project", "", "Restore a deleted project and exit")
		enableGox     = flag.Bool("enable-gox", false, "Enable Gox advanced features (cells, RAG, etc.)")
		goxConfig     = flag.String("gox-config", "config/gox", "Path to Gox configuration directory")
	)

	flag.Parse()

	// Headless mode activates voice + allow-all
	if *headless {
		*enableVoice = true
		*mode = "allow-all"
	}

	// Resolve workbench directory
	workbenchDir, err := os.Getwd()
	if err != nil {
		fatal("Failed to get working directory: %v", err)
	}
	if *workdir != "." {
		workbenchDir = *workdir
	}

	// Initialize project manager
	projectMgr := project.NewManager(workbenchDir)
	if err := projectMgr.EnsureDirectories(); err != nil {
		fatal("Failed to initialize project directories: %v", err)
	}

	// Handle project management commands
	if *listProjects {
		listProjectsCmd(projectMgr)
		return
	}
	if *createProject != "" {
		createProjectCmd(projectMgr, *createProject)
		return
	}
	if *deleteProject != "" {
		deleteProjectCmd(projectMgr, *deleteProject)
		return
	}
	if *restoreProject != "" {
		restoreProjectCmd(projectMgr, *restoreProject)
		return
	}

	// Determine which project to use
	selectedProject := *projectName
	if selectedProject == "" {
		// Load from context manager
		contextMgr := alfacontext.NewManager(workbenchDir)
		selectedProject = contextMgr.GetActiveProject()
	}

	// If still no project, check if there's only one project
	if selectedProject == "" {
		projects, err := projectMgr.List()
		if err != nil {
			fatal("Failed to list projects: %v", err)
		}
		if len(projects) == 0 {
			// No projects exist - prompt user to create the first one
			selectedProject = promptForFirstProject(projectMgr)
		} else if len(projects) == 1 {
			selectedProject = projects[0].Name
			fmt.Printf("Using project: %s\n", selectedProject)
		} else {
			fmt.Println("Multiple projects found. Please specify one with --project:")
			for _, p := range projects {
				fmt.Printf("  - %s (last modified: %s)\n", p.Name, p.LastModified.Format("2006-01-02 15:04"))
			}
			os.Exit(1)
		}
	}

	// Ensure project exists
	if !projectMgr.Exists(selectedProject) {
		fmt.Printf("Project '%s' does not exist. Creating it...\n", selectedProject)
		if err := projectMgr.Create(selectedProject); err != nil {
			fatal("Failed to create project: %v", err)
		}
	}

	// Create VFS instances
	// Workbench VFS: for config, context, history (read/write)
	workbenchVFS, err := vfs.NewVFS(workbenchDir, false)
	if err != nil {
		fatal("Failed to create workbench VFS: %v", err)
	}

	// Project VFS: for source code, sandboxed (read/write)
	projectDir := projectMgr.GetProjectPath(selectedProject)
	projectVFS, err := vfs.NewVFS(projectDir, false)
	if err != nil {
		fatal("Failed to create project VFS: %v", err)
	}

	fmt.Printf("Workbench: %s\n", workbenchVFS.Root())
	fmt.Printf("Project:   %s (%s)\n\n", selectedProject, projectVFS.Root())

	// Determine config file path
	cfgPath := *configFile
	if cfgPath == "" {
		cfgPath = ai.GetConfigPath()
	}

	// Load configuration
	cfg, err := ai.LoadConfigWithEnv(cfgPath)
	if err != nil {
		fatal("Failed to load config: %v", err)
	}

	// Determine which provider to use
	selectedProvider := *provider
	if selectedProvider == "" {
		selectedProvider = cfg.DefaultProvider
	}

	// Validate provider exists in config
	if _, ok := cfg.Providers[selectedProvider]; !ok {
		fatal("Provider '%s' not found in config. Available providers: %v", selectedProvider, getProviderNames(cfg))
	}

	// Create LLM client from config
	llm, err := ai.NewLLMFromConfig(cfg, selectedProvider)
	if err != nil {
		fatal("Failed to create LLM client: %v", err)
	}

	// Setup voice if enabled
	var stt speech.STT
	var tts speech.TTS
	var recorder audio.Recorder
	var player audio.Player

	if *enableVoice {
		openaiKey := os.Getenv("OPENAI_API_KEY")
		if openaiKey == "" {
			fatal("Voice mode requires OPENAI_API_KEY environment variable")
		}

		// Create STT (Whisper)
		stt = speech.NewWhisperSTT(speech.STTConfig{
			APIKey: openaiKey,
			Model:  "whisper-1",
			Timeout: 60 * time.Second,
		})

		// Create TTS
		tts = speech.NewOpenAITTS(speech.TTSConfig{
			APIKey: openaiKey,
			Model:  "tts-1",
			Voice:  "alloy",
			Speed:  1.0,
			Format: "mp3",
			Timeout: 60 * time.Second,
		})

		// Create audio recorder
		recorder = audio.NewVADRecorder(audio.DefaultRecordConfig())
		if !recorder.IsAvailable() {
			fmt.Println("⚠️  Warning: sox not found. Voice input disabled.")
			fmt.Println("   Install with: brew install sox")
			recorder = nil
		}

		// Create audio player
		player = audio.NewSoxPlayer()
		if !player.IsAvailable() {
			fmt.Println("⚠️  Warning: No audio player found. Voice output disabled.")
			player = nil
		}

		if recorder != nil || player != nil {
			fmt.Println("🎤 Voice mode enabled")
		}
	}

	// Setup sandbox if enabled
	var sb sandbox.Sandbox
	if *useSandbox {
		sandboxCfg := sandbox.DefaultConfig()
		sandboxCfg.DefaultImage = *sandboxImage
		sb = sandbox.NewDockerSandbox(sandboxCfg)

		if !sb.IsAvailable() {
			fmt.Println("⚠️  Warning: Docker not available. Sandbox disabled.")
			fmt.Println("   Install Docker Desktop or docker engine.")
			*useSandbox = false
			sb = nil
		} else {
			fmt.Println("🐳 Docker sandbox enabled")
		}
	}

	// Create components
	contextMgr := alfacontext.NewManager(workbenchVFS.Root())
	contextMgr.SetActiveProject(selectedProject)
	toolDispatcher := tools.NewDispatcherWithSandbox(projectVFS, sb, *useSandbox)
	toolDispatcher.SetProjectManager(projectMgr)
	vcrInstance := vcr.NewVcr("assistant", projectVFS.Root())

	// Initialize Gox if enabled
	var goxMgr *gox.Manager
	if *enableGox {
		fmt.Println("🔧 Initializing Gox advanced features...")
		goxMgr, err = gox.NewManager(gox.Config{
			ConfigPath:      *goxConfig,
			DefaultDataRoot: workbenchDir,
			Debug:           true,
		})
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to initialize Gox: %v\n", err)
			fmt.Println("   Advanced features will be disabled.")
			goxMgr = nil
		} else {
			fmt.Println("✅ Gox initialized successfully")
			toolDispatcher.SetGoxManager(goxMgr)
		}
	}

	// Parse mode
	var execMode orchestrator.Mode
	if *mode == "allow-all" {
		execMode = orchestrator.ModeAllowAll
	} else {
		execMode = orchestrator.ModeConfirm
	}

	// Create orchestrator
	orch := orchestrator.New(orchestrator.Config{
		LLM:            llm,
		ContextManager: contextMgr,
		ToolDispatcher: toolDispatcher,
		VCR:            vcrInstance,
		ProjectVFS:     projectVFS,
		ProjectManager: projectMgr,
		WorkbenchRoot:  workbenchDir,
		GoxManager:     goxMgr,
		STT:            stt,
		TTS:            tts,
		Recorder:       recorder,
		Player:         player,
		Mode:           execMode,
		MaxIterations:  *iterations,
	})

	// Cleanup on exit
	if goxMgr != nil {
		defer func() {
			fmt.Println("\n🔧 Shutting down Gox...")
			if err := goxMgr.Close(); err != nil {
				fmt.Printf("⚠️  Warning: Gox shutdown error: %v\n", err)
			}
		}()
	}

	// Run
	ctx := context.Background()
	if err := orch.Run(ctx); err != nil {
		fatal("Error: %v", err)
	}
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Fatal: "+format+"\n", args...)
	os.Exit(1)
}

func getProviderNames(cfg *ai.ConfigFile) []string {
	names := make([]string, 0, len(cfg.Providers))
	for name := range cfg.Providers {
		names = append(names, name)
	}
	return names
}

// Project management commands

func listProjectsCmd(mgr *project.Manager) {
	projects, err := mgr.List()
	if err != nil {
		fatal("Failed to list projects: %v", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found.")
		fmt.Println("\nCreate a new project with:")
		fmt.Println("  alfa --create-project <name>")
		return
	}

	fmt.Printf("Projects (%d):\n\n", len(projects))
	for _, p := range projects {
		fmt.Printf("📁 %s\n", p.Name)
		fmt.Printf("   Path:         %s\n", p.Path)
		fmt.Printf("   Last modified: %s\n", p.LastModified.Format("2006-01-02 15:04:05"))
		if p.IsGitRepo {
			fmt.Printf("   Branch:       %s\n", p.Branch)
			if p.LastCommit != "" {
				commit := p.LastCommit
				if len(commit) > 50 {
					commit = commit[:50] + "..."
				}
				fmt.Printf("   Last commit:  %s\n", commit)
			}
		}
		fmt.Println()
	}
}

func createProjectCmd(mgr *project.Manager, name string) {
	if name == "" {
		fatal("Project name cannot be empty")
	}

	// Validate project name
	if strings.ContainsAny(name, "/\\:*?\"<>|") {
		fatal("Invalid project name. Avoid special characters: /\\:*?\"<>|")
	}

	fmt.Printf("Creating project '%s'...\n", name)
	if err := mgr.Create(name); err != nil {
		fatal("Failed to create project: %v", err)
	}

	meta, _ := mgr.GetMetadata(name)
	fmt.Printf("\n✅ Project created successfully!\n")
	fmt.Printf("   Path: %s\n", meta.Path)
	fmt.Printf("\nStart working on it with:\n")
	fmt.Printf("  alfa --project %s\n", name)
}

func deleteProjectCmd(mgr *project.Manager, name string) {
	if name == "" {
		fatal("Project name cannot be empty")
	}

	if !mgr.Exists(name) {
		fatal("Project '%s' does not exist", name)
	}

	// Confirm deletion
	fmt.Printf("⚠️  Delete project '%s'?\n", name)
	fmt.Println("   The project files will be deleted, but a backup is kept in .git-remotes/")
	fmt.Println("   You can restore it later with: alfa --restore-project " + name)
	fmt.Print("\n   Type 'yes' to confirm: ")

	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "yes" {
		fmt.Println("Cancelled.")
		return
	}

	if err := mgr.Delete(name); err != nil {
		fatal("Failed to delete project: %v", err)
	}

	fmt.Printf("\n✅ Project '%s' deleted.\n", name)
	fmt.Printf("   Restore with: alfa --restore-project %s\n", name)
}

func restoreProjectCmd(mgr *project.Manager, name string) {
	if name == "" {
		fatal("Project name cannot be empty")
	}

	fmt.Printf("Restoring project '%s'...\n", name)
	if err := mgr.Restore(name); err != nil {
		fatal("Failed to restore project: %v", err)
	}

	meta, _ := mgr.GetMetadata(name)
	fmt.Printf("\n✅ Project restored successfully!\n")
	fmt.Printf("   Path: %s\n", meta.Path)
	fmt.Printf("\nStart working on it with:\n")
	fmt.Printf("  alfa --project %s\n", name)
}

func promptForFirstProject(mgr *project.Manager) string {
	fmt.Println("🎉 Welcome to Alfa!")
	fmt.Println()
	fmt.Println("No projects found in this workbench.")
	fmt.Println("Let's create your first project to get started.")
	fmt.Println()
	fmt.Print("Enter a name for your first project: ")

	var projectName string
	fmt.Scanln(&projectName)

	// Trim whitespace
	projectName = strings.TrimSpace(projectName)

	// Validate
	if projectName == "" {
		fmt.Println("\n❌ Project name cannot be empty.")
		fmt.Println("You can create a project later with: alfa --create-project <name>")
		os.Exit(1)
	}

	if strings.ContainsAny(projectName, "/\\:*?\"<>|") {
		fmt.Println("\n❌ Invalid project name. Avoid special characters: /\\:*?\"<>|")
		fmt.Println("You can create a project later with: alfa --create-project <name>")
		os.Exit(1)
	}

	// Create the project
	fmt.Printf("\nCreating project '%s'...\n", projectName)
	if err := mgr.Create(projectName); err != nil {
		fatal("Failed to create project: %v", err)
	}

	fmt.Printf("✅ Project '%s' created successfully!\n\n", projectName)

	return projectName
}
