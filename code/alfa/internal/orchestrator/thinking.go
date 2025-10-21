package orchestrator

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ThinkingIndicator shows an animated "thinking" message with tips and progress
// Similar to Claude Code's multi-line indicator that disappears when done
type ThinkingIndicator struct {
	verb      string
	tip       string
	phase     string // Current PEV phase: "Planning", "Executing", "Verifying"
	detail    string // Current action detail
	stopChan  chan struct{}
	wg        sync.WaitGroup
	lineCount int         // Number of lines printed (for clearing)
	mu        sync.Mutex  // Protects phase and detail
}

var thinkingVerbs = []string{
	"Pondering",
	"Ruminating",
	"Cogitating",
	"Mulling",
	"Contemplating",
	"Deliberating",
	"Percolating",
	"Marinating",
	"Brewing",
	"Noodling",
	"Tinkering",
	"Fiddling",
	"Scrutinizing",
	"Investigating",
	"Probing",
	"Excavating",
	"Unearthing",
	"Mining",
	"Spelunking",
	"Deciphering",
	"Untangling",
	"Unraveling",
	"Unpacking",
	"Dissecting",
	"Parsing",
	"Wrangling",
	"Wrestling",
	"Grappling",
	"Massaging",
	"Kneading",
	"Synthesizing",
	"Distilling",
	"Crystallizing",
	"Formulating",
	"Concocting",
	"Hatching",
	"Incubating",
	"Germinating",
	"Cultivating",
	"Crafting",
	"Fashioning",
	"Forging",
	"Smithing",
	"Sculpting",
	"Chiseling",
	"Weaving",
	"Knitting",
	"Stitching",
	"Assembling",
	"Orchestrating",
	"Choreographing",
	"Conducting",
	"Harmonizing",
	"Tuning",
	"Calibrating",
	"Balancing",
	"Juggling",
	"Navigating",
	"Charting",
	"Mapping",
	"Surveying",
	"Scouting",
	"Reconnoitering",
	"Exploring",
	"Traversing",
	"Meandering",
	"Wandering",
	"Roaming",
	"Sifting",
	"Combing",
	"Raking",
	"Trawling",
	"Dredging",
	"Harvesting",
	"Gleaning",
	"Gathering",
	"Collating",
	"Aggregating",
	"Consolidating",
	"Integrating",
	"Fusing",
	"Melding",
	"Blending",
	"Merging",
	"Converging",
	"Coalescing",
	"Perusing",
	"Scanning",
	"Skimming",
	"Absorbing",
	"Digesting",
	"Metabolizing",
	"Assimilating",
	"Processing",
	"Crunching",
	"Grinding",
	"Milling",
	"Refining",
	"Polishing",
	"Buffing",
}

var helpfulTips = []string{
	"Tip: Type 'clear' to reset conversation context",
	"Tip: Press Ctrl+C to interrupt at any time",
	"Tip: Session logs are saved in the logs/ directory",
	"Tip: Use 'exit' or 'quit' to end the session",
	"Tip: Multi-line input supported - press Ctrl+D to submit",
	"Tip: Check workbench/config/alfa.yaml for settings",
	"Tip: Self-modification mode lets Alfa modify its own code",
	"Tip: Enable voice with voice.input_enabled in config",
	"Tip: Projects are isolated in workbench/projects/",
	"Tip: Use config_list to see all current settings",
}

// NewThinkingIndicator creates a new thinking indicator with random verb and tip
func NewThinkingIndicator() *ThinkingIndicator {
	verb := thinkingVerbs[rand.Intn(len(thinkingVerbs))]
	tip := helpfulTips[rand.Intn(len(helpfulTips))]
	return &ThinkingIndicator{
		verb:      verb,
		tip:       tip,
		phase:     "",
		detail:    "",
		stopChan:  make(chan struct{}),
		lineCount: 2, // We print 2 lines
	}
}

// UpdateProgress updates the current phase and detail (for PEV progress tracking)
func (t *ThinkingIndicator) UpdateProgress(phase, detail string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.phase = phase
	t.detail = detail
}

// Start begins the animated thinking indicator
// Displays multi-line output like Claude Code:
//   ⠋ Pondering… (Ctrl+C to interrupt)
//     ⎿ Planning: Analyzing request...
func (t *ThinkingIndicator) Start() {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()

		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		idx := 0
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		// Print initial multi-line message
		fmt.Printf("%s %s… (Ctrl+C to interrupt)\n", spinner[idx], t.verb)
		t.mu.Lock()
		secondLine := t.getSecondLine()
		t.mu.Unlock()
		fmt.Printf("  ⎿ %s\n", secondLine)

		for {
			select {
			case <-t.stopChan:
				return
			case <-ticker.C:
				idx = (idx + 1) % len(spinner)

				// Get current progress info
				t.mu.Lock()
				secondLine := t.getSecondLine()
				t.mu.Unlock()

				// Move cursor up 2 lines, clear them, and redraw
				// ANSI: \033[2A = move up 2 lines, \r = start of line, \033[K = clear to end
				fmt.Printf("\033[2A\r\033[K%s %s… (Ctrl+C to interrupt)\n", spinner[idx], t.verb)
				fmt.Printf("\r\033[K  ⎿ %s\n", secondLine)
			}
		}
	}()
}

// getSecondLine returns the appropriate second line (must be called with lock held)
func (t *ThinkingIndicator) getSecondLine() string {
	if t.phase != "" {
		// Show PEV progress
		if t.detail != "" {
			return fmt.Sprintf("%s: %s", t.phase, t.detail)
		}
		return fmt.Sprintf("%s…", t.phase)
	}
	// Show tip
	return t.tip
}

// Stop stops the animation and completely clears all lines
// The response will appear in the same spot (Claude Code style)
func (t *ThinkingIndicator) Stop() {
	close(t.stopChan)
	t.wg.Wait()

	// Move cursor up to start of thinking indicator and clear everything below
	// ANSI codes:
	//   \033[2A - move up 2 lines (to start of indicator)
	//   \r - carriage return (start of line)
	//   \033[J - clear from cursor to end of screen
	fmt.Printf("\033[2A\r\033[J")

	// Cursor is now at the start of the line where thinking indicator was
	// Response will appear in the exact same spot
}
