package eventbus

import (
	"slices"
	"sync"
)

// Handler defines the signature for event handlers.
type Handler func(data any)

// subscriber stores a handler and a unique id for safe comparison.
type subscriber struct {
	id int
	h  Handler
}

// EventBus struct supports event subscription and publishing.
type EventBus struct {
	subscribers map[string][]subscriber
	mu          sync.RWMutex
	nextID      int
}

// New creates a new EventBus instance.
func New() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]subscriber),
	}
}

// Subscribe adds a handler for a named event and returns its id.
func (eb *EventBus) Subscribe(event string, h Handler) int {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.nextID++
	id := eb.nextID

	eb.subscribers[event] = append(eb.subscribers[event], subscriber{id, h})
	return id
}

// Unsubscribe removes a handler by id for a specific event.
func (eb *EventBus) Unsubscribe(event string, id int) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subs, exists := eb.subscribers[event]
	if !exists {
		return
	}

	for i, sub := range subs {
		if sub.id == id {
			subs = append(subs[:i], subs[i+1:]...)
			break
		}
	}

	// Remove event key if no more subscribers
	if len(subs) == 0 {
		delete(eb.subscribers, event)
		return
	}

	eb.subscribers[event] = subs
}

// Publish triggers all handlers for a named event.
func (eb *EventBus) Publish(event string, data any) {
	eb.mu.RLock()
	subs := slices.Clone(eb.subscribers[event])
	eb.mu.RUnlock()

	for _, subs := range subs {
		subs.h(data)
	}
}
