package sse

import (
	"sync"
)

// Event represents a server-sent event to push to connected clients.
type Event struct {
	Type    string `json:"type"`
	EmailID string `json:"emailId"`
}

// Broker manages per-user SSE subscriptions.
type Broker struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan Event]struct{}
}

// NewBroker creates a new SSE broker.
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string]map[chan Event]struct{}),
	}
}

// Subscribe registers a new listener for the given user. Returns the event
// channel and an unsubscribe function that must be called when done.
func (b *Broker) Subscribe(userID string) (chan Event, func()) {
	ch := make(chan Event, 8)

	b.mu.Lock()
	if b.subscribers[userID] == nil {
		b.subscribers[userID] = make(map[chan Event]struct{})
	}
	b.subscribers[userID][ch] = struct{}{}
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		delete(b.subscribers[userID], ch)
		if len(b.subscribers[userID]) == 0 {
			delete(b.subscribers, userID)
		}
		b.mu.Unlock()
		close(ch)
	}

	return ch, unsubscribe
}

// Publish sends an event to all subscribers for the given user.
// Non-blocking: if a client's channel is full, the event is dropped.
func (b *Broker) Publish(userID string, event Event) {
	b.mu.RLock()
	subs := b.subscribers[userID]
	b.mu.RUnlock()

	for ch := range subs {
		select {
		case ch <- event:
		default:
			// Client too slow, drop the event
		}
	}
}
