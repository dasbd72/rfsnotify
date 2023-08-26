package rfsnotify

import (
	"sync"

	"github.com/fsnotify/fsnotify"
)

type eventQueue struct {
	mu     sync.Mutex
	events []fsnotify.Event
}

func newEventQueue() *eventQueue {
	return &eventQueue{
		events: make([]fsnotify.Event, 0),
	}
}

func (q *eventQueue) size() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.events)
}

func (q *eventQueue) push(event fsnotify.Event) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.events = append(q.events, event)
}

func (q *eventQueue) pop() fsnotify.Event {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.events) == 0 {
		return fsnotify.Event{}
	}

	event := q.events[0]
	q.events = q.events[1:]

	return event
}
