package orchestrator

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// EventBridge translates Gox broker messages to Go channels
// This allows host applications (like Alfa) to subscribe to Gox events
// using native Go channels instead of the broker protocol
//
// NOTE: This is a simplified Phase 1 implementation.
// Phase 2 will add actual broker integration.
//
// Current: In-memory event forwarding only
// Future: Full broker message translation
type EventBridge struct {
	subscribers map[string][]chan Event
	mutex       sync.RWMutex
}

// Initialize initializes the EventBridge
func (eb *EventBridge) Initialize() {
	eb.subscribers = make(map[string][]chan Event)
}

// Subscribe returns a channel that receives events for the given topic pattern
func (eb *EventBridge) Subscribe(topicPattern string) <-chan Event {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	// Create buffered channel to prevent blocking
	ch := make(chan Event, 100)

	// Add to subscribers
	eb.subscribers[topicPattern] = append(eb.subscribers[topicPattern], ch)

	return ch
}

// Unsubscribe closes a subscription channel
func (eb *EventBridge) Unsubscribe(topicPattern string, ch <-chan Event) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	subscribers := eb.subscribers[topicPattern]
	newSubscribers := []chan Event{}

	for _, subscriber := range subscribers {
		// Compare using unsafe pointer conversion
		// This is safe because we're only comparing, not using the pointers
		if fmt.Sprintf("%p", subscriber) != fmt.Sprintf("%p", ch) {
			newSubscribers = append(newSubscribers, subscriber)
		} else {
			// Found the channel to unsubscribe
			close(subscriber)
		}
	}

	eb.subscribers[topicPattern] = newSubscribers
}

// TopicMatches checks if a topic matches a pattern (public for testing)
func (eb *EventBridge) TopicMatches(topic, pattern string) bool {
	return eb.topicMatches(topic, pattern)
}

// topicMatches checks if a topic matches a pattern
func (eb *EventBridge) topicMatches(topic, pattern string) bool {
	if pattern == "*" {
		return true // Match all
	}

	// Split topic and pattern
	topicParts := strings.Split(topic, ":")
	patternParts := strings.Split(pattern, ":")

	// Different number of parts - no match
	if len(topicParts) != len(patternParts) {
		return false
	}

	// Check each part
	for i := range topicParts {
		if patternParts[i] == "*" {
			continue // Wildcard matches anything
		}
		if topicParts[i] != patternParts[i] {
			return false // Mismatch
		}
	}

	return true
}

// Publish publishes an event to a topic
func (eb *EventBridge) Publish(topic string, data interface{}) error {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	// Create event
	event := Event{
		Topic:     topic,
		ProjectID: eb.extractProjectID(topic),
		Timestamp: time.Now(),
		Source:    "host_application",
	}

	// Convert data to map
	if dataMap, ok := data.(map[string]interface{}); ok {
		event.Data = dataMap
	} else {
		event.Data = map[string]interface{}{"payload": data}
	}

	// Forward to matching subscribers
	for pattern, subscribers := range eb.subscribers {
		if eb.topicMatches(topic, pattern) {
			for _, subscriber := range subscribers {
				// Non-blocking send
				select {
				case subscriber <- event:
				default:
				}
			}
		}
	}

	return nil
}

// PublishAndWait publishes a request and waits for a response
func (eb *EventBridge) PublishAndWait(
	requestTopic string,
	responseTopic string,
	data interface{},
	timeout time.Duration,
) (*Event, error) {
	// Subscribe to response topic
	responseCh := eb.Subscribe(responseTopic)
	defer eb.Unsubscribe(responseTopic, responseCh)

	// Publish request
	if err := eb.Publish(requestTopic, data); err != nil {
		return nil, fmt.Errorf("failed to publish request: %w", err)
	}

	// Wait for response
	select {
	case event := <-responseCh:
		return &event, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for response on topic %s", responseTopic)
	}
}

// extractProjectID extracts project ID from topic
func (eb *EventBridge) extractProjectID(topic string) string {
	if strings.Contains(topic, ":") {
		parts := strings.SplitN(topic, ":", 2)
		return parts[0]
	}
	return "default"
}

// Close closes all subscription channels
func (eb *EventBridge) Close() {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	for pattern, subscribers := range eb.subscribers {
		for _, subscriber := range subscribers {
			close(subscriber)
		}
		delete(eb.subscribers, pattern)
	}
}
