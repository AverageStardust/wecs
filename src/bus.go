package main

import (
	"iter"

	"github.com/averagestardust/wecs/internal/ring"
)

// A generic bus to pass events around the application.
type Bus[Event any] struct {
	listeners  []func(event Event)
	pipes      []*Pipe[Event]
	eventQueue *ring.Ring[Event]
}

// A pipe that can consume events from a bus on it's own time.
type Pipe[Event any] struct {
	bus       *Bus[Event]
	nextEvent uint64
}

// Create an event bus to pass events around the application.
func NewBus[Event any]() Bus[Event] {
	return Bus[Event]{
		listeners:  nil,
		pipes:      nil,
		eventQueue: ring.NewRing[Event](),
	}
}

// Create a pipe that can consume events from a bus on it's own time.
// Once a bus has a pipe it must queue events until all pipes have consumed them.
func (bus *Bus[Event]) NewPipe() *Pipe[Event] {
	return &Pipe[Event]{
		bus:       bus,
		nextEvent: bus.eventQueue.Head(),
	}
}

// Add a listener to a bus that immediately is called when events are published.
func (bus *Bus[Event]) Listen(listener func(event Event)) {
	bus.listeners = append(bus.listeners, listener)
}

// Send a event over the bus
func (bus *Bus[Event]) Publish(event Event) {
	for _, subscriber := range bus.listeners {
		subscriber(event)
	}

	if len(bus.pipes) > 0 {
		bus.eventQueue.Enqueue(event)
	}
}

// Send multiple events over the bus
func (bus *Bus[Event]) PublishBatch(events []Event) {
	for _, subscriber := range bus.listeners {
		for _, event := range events {
			subscriber(event)
		}
	}

	if len(bus.pipes) > 0 {
		bus.eventQueue.EnqueueBatch(events)
	}
}

// Delete events that have been consumed by all pipes on the bus.
func (bus *Bus[Event]) dropConsumedQueue() {
	lastAccessibleEvent := bus.pipes[0].nextEvent
	for _, pipe := range bus.pipes[1:] {
		lastAccessibleEvent = min(lastAccessibleEvent, pipe.nextEvent)
	}

	bus.eventQueue.DropUntil(lastAccessibleEvent)
}

// Get an iterator of events queued on a pipe.
func (pipe *Pipe[Event]) Iter() iter.Seq[Event] {
	return func(yield func(Event) bool) {
		for {
			event, success := pipe.bus.eventQueue.Peek(pipe.nextEvent)
			if !success {
				break
			}

			pipe.nextEvent++
			pipe.bus.dropConsumedQueue()
			if !yield(event) {
				break
			}
		}
	}
}

// Get one event from a pipe.
func (pipe *Pipe[Event]) Pop() (event Event, success bool) {
	event, success = pipe.bus.eventQueue.Peek(pipe.nextEvent)

	if success {
		pipe.nextEvent++
		pipe.bus.dropConsumedQueue()
	}

	return
}

// Close a pipe when it is no longer needed.
// This frees a bus to free unconsumed events.
func (pipe *Pipe[Event]) Close() {
	// find index
	var index int
	for i, otherPipe := range pipe.bus.pipes {
		if otherPipe == pipe {
			index = i
			break
		}
	}

	// shift all down, overwriting pipe
	for i := index + 1; i < len(pipe.bus.pipes); i++ {
		pipe.bus.pipes[i-1] = pipe.bus.pipes[i]
	}

	// drop last pipe
	pipe.bus.pipes = pipe.bus.pipes[0 : len(pipe.bus.pipes)-1]
	pipe.bus = nil
}
