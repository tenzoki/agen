package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tenzoki/agen/alfa/internal/ai"
	"github.com/tenzoki/agen/alfa/internal/audio"
	alfaconfig "github.com/tenzoki/agen/alfa/internal/config"
	alfacontext "github.com/tenzoki/agen/alfa/internal/context"
	"github.com/tenzoki/agen/alfa/internal/knowledge"
	"github.com/tenzoki/agen/alfa/internal/orchestrator"
	"github.com/tenzoki/agen/alfa/internal/project"
	"github.com/tenzoki/agen/alfa/internal/sandbox"
	"github.com/tenzoki/agen/alfa/internal/speech"
	"github.com/tenzoki/agen/alfa/internal/tools"
	"github.com/tenzoki/agen/atomic/vcr"
	"github.com/tenzoki/agen/atomic/vfs"
	cellorchestrator "github.com/tenzoki/agen/cellorg/public/orchestrator"
)

func main() {
	var (
		workbench         = flag.String("workbench", "", "Workbench directory (default: workbench)")
		configFile        = flag.String("config", "", "AI config file path (default: workbench/config/ai-config.json)")
		autoConfirm       = flag.Bool("auto-confirm", false, "Auto-confirm all operations")
		provider          = flag.String("provider", "", "AI provider override (anthropic or openai)")
		iterations        = flag.Int("max-iterations", 0, "Maximum AI iterations per request")
		enableVoice       = flag.Bool("voice", false, "Enable voice input/output (requires OPENAI_API_KEY and sox)")
		headless          = flag.Bool("headless", false, "Headless mode: enables voice + auto-confirm")
		useSandbox        = flag.Bool("sandbox", false, "Use Docker sandbox for command execution")
		sandboxImage      = flag.String("sandbox-image", "", "Docker image for sandbox")
		projectName       = flag.String("project", "", "Project name (creates if doesn't exist)")
		listProjects      = flag.Bool("list-projects", false, "List all projects and exit")
		createProject     = flag.String("create-project", "", "Create a new project and exit")
		deleteProject     = flag.String("delete-project", "", "Delete a project (keeps backup) and exit")
		restoreProject    = flag.String("restore-project", "", "Restore a deleted project and exit")
		enableCellorg     = flag.Bool("enable-cellorg", false, "Enable cellorg advanced features")
		cellorgConfig     = flag.String("cellorg-config", "", "Path to cellorg configuration directory")
		captureOutput     = flag.Bool("capture-output", true, "Capture command output to show AI")
		maxOutputKB       = flag.Int("max-output", 0, "Maximum output size in KB to show AI")
		allowSelfModify   = flag.Bool("allow-self-modify", false, "Allow AI to modify framework code")
	)

	flag.Parse()

	// Determine workbench directory (CLI has priority)
	workbenchDir := determineWorkbenchPath(*workbench)

	// Load or create alfa.yaml configuration (non-fatal, use defaults on error)
	cfg, err := alfaconfig.LoadOrCreate(workbenchDir)
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to load configuration: %v\n", err)
		fmt.Printf("    Using default configuration instead.\n\n")
		defaultCfg := alfaconfig.DefaultConfig()
		defaultCfg.Workbench.Path = workbenchDir
		cfg = &defaultCfg
	}

	// Apply CLI overrides (CLI has priority over alfa.yaml)
	cliFlags := alfaconfig.CLIFlags{
		Workbench:       *workbench,
		ConfigFile:      *configFile,
		AutoConfirm:     *autoConfirm,
		Provider:        *provider,
		MaxIterations:   *iterations,
		Voice:           *enableVoice,
		Headless:        *headless,
		Sandbox:         *useSandbox,
		SandboxImage:    *sandboxImage,
		Project:         *projectName,
		EnableCellorg:   *enableCellorg,
		CellorgConfig:   *cellorgConfig,
		CaptureOutput:   captureOutput,
		MaxOutputKB:     *maxOutputKB,
		AllowSelfModify: *allowSelfModify,
	}
	cfg.ApplyCLIOverrides(cliFlags)

	// Update workbench path in case it was changed by CLI
	workbenchDir = cfg.Workbench.Path

	// Try to save the final configuration back to alfa.yaml (non-fatal)
	configPath := alfaconfig.ResolveConfigPath(workbenchDir)
	if err := alfaconfig.SaveConfig(configPath, cfg); err != nil {
		fmt.Printf("⚠️  Warning: Failed to save configuration: %v\n", err)
	}

	// Extract knowledge base if needed
	frameworkRoot := filepath.Dir(workbenchDir)
	extractor := knowledge.NewExtractor(frameworkRoot)
	if extractor.IsStale() {
		if err := extractor.Extract(); err != nil {
			fmt.Printf("⚠️  Warning: Knowledge extraction failed: %v\n", err)
		}
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
		deleteProjectCmd(projectMgr, workbenchDir, *deleteProject)
		return
	}
	if *restoreProject != "" {
		restoreProjectCmd(projectMgr, *restoreProject)
		return
	}

	// Determine which project to use (from config or context)
	selectedProject := cfg.Workbench.Project
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

	// Load AI configuration (either from ai-config.json or use embedded config)
	var aiCfg *ai.ConfigFile
	aiConfigPath := cfg.AI.ConfigFile
	if aiConfigPath == "" {
		aiConfigPath = filepath.Join(workbenchDir, "config", "ai-config.json")
	}

	// Try loading from file, fallback to embedded config
	aiCfg, err = ai.LoadConfigWithEnv(aiConfigPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Use embedded config from alfa.yaml
			aiCfg = convertToAIConfig(cfg)
		} else {
			fatal("Failed to load AI config: %v", err)
		}
	}

	// Override provider if specified
	selectedProvider := cfg.AI.Provider
	if selectedProvider == "" {
		selectedProvider = aiCfg.DefaultProvider
	}

	// Validate provider exists in config (fallback to first available if not found)
	if _, ok := aiCfg.Providers[selectedProvider]; !ok {
		fmt.Printf("⚠️  Warning: Provider '%s' not found in config.\n", selectedProvider)
		fmt.Printf("    Available providers: %v\n", getProviderNames(aiCfg))

		// Try to use first available provider
		if len(aiCfg.Providers) > 0 {
			for name := range aiCfg.Providers {
				selectedProvider = name
				fmt.Printf("    Using '%s' instead.\n\n", selectedProvider)
				break
			}
		} else {
			fatal("No AI providers configured. Please set ANTHROPIC_API_KEY or OPENAI_API_KEY")
		}
	}

	// Create LLM client from config
	llm, err := ai.NewLLMFromConfig(aiCfg, selectedProvider)
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to create LLM client: %v\n", err)
		fmt.Printf("    Please check your API key environment variables.\n")
		fatal("Cannot continue without a valid AI provider")
	}

	// Debug: Show which model is being used
	fmt.Printf("🤖 Using: %s (%s)\n", llm.Model(), llm.Provider())

	// Setup voice if enabled
	var stt speech.STT
	var tts speech.TTS
	var recorder audio.Recorder
	var player audio.Player

	if cfg.Voice.InputEnabled || cfg.Voice.OutputEnabled {
		openaiKey := os.Getenv("OPENAI_API_KEY")
		if openaiKey == "" {
			fatal("Voice mode requires OPENAI_API_KEY environment variable")
		}

		// Create STT (Whisper) if input enabled
		if cfg.Voice.InputEnabled {
			stt = speech.NewWhisperSTT(speech.STTConfig{
				APIKey:  openaiKey,
				Model:   "whisper-1",
				Timeout: 60 * time.Second,
			})

			// Create audio recorder
			recorder = audio.NewVADRecorder(audio.DefaultRecordConfig())
			if !recorder.IsAvailable() {
				fmt.Println("⚠️  Warning: sox not found. Voice input disabled.")
				fmt.Println("   Install with: brew install sox")
				recorder = nil
				stt = nil
			}
		}

		// Create TTS if output enabled
		if cfg.Voice.OutputEnabled {
			tts = speech.NewOpenAITTS(speech.TTSConfig{
				APIKey:  openaiKey,
				Model:   "tts-1",
				Voice:   "alloy",
				Speed:   1.0,
				Format:  "mp3",
				Timeout: 60 * time.Second,
			})

			// Create audio player
			player = audio.NewSoxPlayer()
			if !player.IsAvailable() {
				fmt.Println("⚠️  Warning: No audio player found. Voice output disabled.")
				player = nil
				tts = nil
			}
		}

		if stt != nil || tts != nil {
			fmt.Println("🎤 Voice mode enabled")
			if stt != nil {
				fmt.Println("   ✓ Voice input active")
			}
			if tts != nil {
				fmt.Println("   ✓ Voice output active")
			}
		}
	}

	// Setup sandbox if enabled
	var sb sandbox.Sandbox
	if cfg.Sandbox.Enabled {
		sandboxCfg := sandbox.DefaultConfig()
		sandboxCfg.DefaultImage = cfg.Sandbox.Image
		sb = sandbox.NewDockerSandbox(sandboxCfg)

		if !sb.IsAvailable() {
			fmt.Println("⚠️  Warning: Docker not available. Sandbox disabled.")
			fmt.Println("   Install Docker Desktop or docker engine.")
			cfg.Sandbox.Enabled = false
			sb = nil
		} else {
			fmt.Println("🐳 Docker sandbox enabled")
		}
	}

	// Create components
	contextMgr := alfacontext.NewManager(workbenchVFS.Root())
	contextMgr.SetActiveProject(selectedProject)
	toolDispatcher := tools.NewDispatcherWithSandbox(projectVFS, sb, cfg.Sandbox.Enabled)
	toolDispatcher.SetProjectManager(projectMgr)
	toolDispatcher.SetOutputCapture(cfg.Output.CaptureEnabled, cfg.Output.MaxSizeKB*1024) // Convert KB to bytes
	toolDispatcher.SetConfig(cfg, configPath) // Enable runtime config management
	vcrInstance := vcr.NewVcr("assistant", projectVFS.Root())

	// Initialize Cellorg if enabled
	var cellMgr *cellorchestrator.EmbeddedOrchestrator
	if cfg.Cellorg.Enabled {
		fmt.Println("🔧 Initializing cellorg advanced features...")
		cellMgr, err = cellorchestrator.NewEmbedded(cellorchestrator.Config{
			ConfigPath:      cfg.Cellorg.ConfigPath,
			DefaultDataRoot: workbenchDir,
			Debug:           true,
		})
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to initialize cellorg: %v\n", err)
			fmt.Println("   Advanced features will be disabled.")
			cellMgr = nil
		} else {
			fmt.Println("✅ Cellorg initialized successfully")
			toolDispatcher.SetCellManager(cellMgr)
		}
	}

	// Parse mode
	var execMode orchestrator.Mode
	if cfg.Execution.AutoConfirm {
		execMode = orchestrator.ModeAutoConfirm
	} else {
		execMode = orchestrator.ModeConfirm
	}

	// Create orchestrator
	orch := orchestrator.New(orchestrator.Config{
		LLM:             llm,
		ContextManager:  contextMgr,
		ToolDispatcher:  toolDispatcher,
		VCR:             vcrInstance,
		ProjectVFS:      projectVFS,
		ProjectManager:  projectMgr,
		WorkbenchRoot:   workbenchDir,
		CellManager:     cellMgr,
		STT:             stt,
		TTS:             tts,
		Recorder:        recorder,
		Player:          player,
		Mode:            execMode,
		MaxIterations:   cfg.Execution.MaxIterations,
		AllowSelfModify: cfg.SelfModify.Allowed,
	})

	// Set voice controller for runtime voice management
	toolDispatcher.SetVoiceController(orch)

	// Cleanup on exit
	if cellMgr != nil {
		defer func() {
			fmt.Println("\n🔧 Shutting down cellorg...")
			if err := cellMgr.Close(); err != nil {
				fmt.Printf("⚠️  Warning: Cellorg shutdown error: %v\n", err)
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

// determineWorkbenchPath resolves the workbench directory path
// Priority: CLI arg > default "workbench" in current directory
func determineWorkbenchPath(cliWorkbench string) string {
	if cliWorkbench != "" {
		// Absolute or relative path provided
		if filepath.IsAbs(cliWorkbench) {
			return cliWorkbench
		}
		// Make absolute
		absPath, err := filepath.Abs(cliWorkbench)
		if err != nil {
			fatal("Failed to resolve workbench path: %v", err)
		}
		return absPath
	}

	// Default: workbench in current directory
	pwd, err := os.Getwd()
	if err != nil {
		fatal("Failed to get working directory: %v", err)
	}
	return filepath.Join(pwd, "workbench")
}

// convertToAIConfig converts alfa config to AI config format
func convertToAIConfig(cfg *alfaconfig.AlfaConfig) *ai.ConfigFile {
	aiCfg := &ai.ConfigFile{
		DefaultProvider: cfg.AI.Provider,
		Providers:       make(map[string]ai.Config),
	}

	// For each provider, get the model to use and its config
	for name, provider := range cfg.AI.Providers {
		// Determine which model to use for this provider
		modelName := provider.DefaultModel
		if name == cfg.AI.Provider && cfg.AI.SelectedModel != "" {
			// If this is the active provider and a specific model is selected, use it
			modelName = cfg.AI.SelectedModel
		}

		// Get the model's configuration
		modelCfg, ok := provider.Models[modelName]
		if !ok {
			// Fallback to first available model if default not found
			for firstModel, firstCfg := range provider.Models {
				modelName = firstModel
				modelCfg = firstCfg
				break
			}
		}

		aiCfg.Providers[name] = ai.Config{
			Model:       modelName,
			MaxTokens:   modelCfg.MaxTokens,
			Temperature: modelCfg.Temperature,
			Timeout:     modelCfg.Timeout,
			RetryCount:  modelCfg.RetryCount,
			RetryDelay:  modelCfg.RetryDelay,
		}
	}

	// Merge with environment variables for API keys
	if anthropicKey := os.Getenv("ANTHROPIC_API_KEY"); anthropicKey != "" {
		if providerCfg, ok := aiCfg.Providers["anthropic"]; ok {
			providerCfg.APIKey = anthropicKey
			aiCfg.Providers["anthropic"] = providerCfg
		}
	}

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		if providerCfg, ok := aiCfg.Providers["openai"]; ok {
			providerCfg.APIKey = openaiKey
			aiCfg.Providers["openai"] = providerCfg
		}
	}

	return aiCfg
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

func deleteProjectCmd(mgr *project.Manager, workbenchDir string, name string) {
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

	// Check if we're deleting the active project - clear context if so
	contextMgr := alfacontext.NewManager(workbenchDir)
	activeProject := contextMgr.GetActiveProject()

	if err := mgr.Delete(name); err != nil {
		fatal("Failed to delete project: %v", err)
	}

	// Clear context if we deleted the active project
	// SetActiveProject automatically saves to disk
	if activeProject == name {
		contextMgr.SetActiveProject("")
	}

	fmt.Printf("\n✅ Project '%s' deleted.\n", name)
	if activeProject == name {
		fmt.Println("   This was your active project. Context has been cleared.")
		fmt.Println("   Use --project to select another project next time you start alfa.")
	}
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
