package keyboard

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/term"
)

// Listener listens for keyboard input in the background
type Listener struct {
	stopChan   chan struct{}
	escHandler func()
	mu         sync.Mutex
	running    bool
	oldState   *term.State
}

// NewListener creates a new keyboard listener
func NewListener() *Listener {
	return &Listener{
		stopChan: make(chan struct{}),
	}
}

// OnEscape sets the handler for ESC key press
func (l *Listener) OnEscape(handler func()) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.escHandler = handler
}

// Start begins listening for keyboard input
func (l *Listener) Start() error {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return nil
	}
	l.running = true
	l.mu.Unlock()

	// Put terminal into raw mode
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	l.oldState = oldState

	go l.listen()
	return nil
}

// Stop stops the keyboard listener
func (l *Listener) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.running {
		return
	}

	close(l.stopChan)
	l.running = false

	// Restore terminal state
	if l.oldState != nil {
		fd := int(os.Stdin.Fd())
		term.Restore(fd, l.oldState)
		l.oldState = nil
	}
}

// listen is the background goroutine that reads keyboard input
func (l *Listener) listen() {
	buf := make([]byte, 3)

	for {
		select {
		case <-l.stopChan:
			return
		default:
			// Read with non-blocking check
			n, err := os.Stdin.Read(buf)
			if err != nil {
				continue
			}

			if n > 0 {
				// ESC key is ASCII 27 (0x1B)
				if buf[0] == 27 {
					l.mu.Lock()
					handler := l.escHandler
					l.mu.Unlock()

					if handler != nil {
						handler()
					}
				}
			}
		}
	}
}
