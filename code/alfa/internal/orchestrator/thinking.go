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

// ThinkingIndicator shows an animated "thinking" message
type ThinkingIndicator struct {
	verb     string
	stopChan chan struct{}
	wg       sync.WaitGroup
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

// NewThinkingIndicator creates a new thinking indicator with a random verb
func NewThinkingIndicator() *ThinkingIndicator {
	verb := thinkingVerbs[rand.Intn(len(thinkingVerbs))]
	return &ThinkingIndicator{
		verb:     verb,
		stopChan: make(chan struct{}),
	}
}

// Start begins the blinking animation
func (t *ThinkingIndicator) Start() {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()

		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		idx := 0
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		// Print initial message
		fmt.Printf("\r%s %s...   ", spinner[idx], t.verb)

		for {
			select {
			case <-t.stopChan:
				return
			case <-ticker.C:
				idx = (idx + 1) % len(spinner)
				// Use \r to go back to start of line and overwrite
				fmt.Printf("\r%s %s...   ", spinner[idx], t.verb)
			}
		}
	}()
}

// Stop stops the animation and clears the line
func (t *ThinkingIndicator) Stop() {
	close(t.stopChan)
	t.wg.Wait()

	// Clear the line by printing spaces and carriage return
	fmt.Printf("\r%s\r", "                                                  ")
}
