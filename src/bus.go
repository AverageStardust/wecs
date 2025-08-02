package main

import (
	"iter"

	"github.com/averagestardust/wecs/internal/ring"
)

type Bus[T any] struct {
	listeners  []func(event T)
	pipes      []*Pipe[T]
	eventQueue ring.Ring[T]
}

type Pipe[T any] struct {
	bus       *Bus[T]
	nextEvent uint64
}

func NewBus[T any]() Bus[T] {
	return Bus[T]{
		listeners:  []func(event T){},
		pipes:      []*Pipe[T]{},
		eventQueue: ring.Ring[T]{},
	}
}

func (bus *Bus[T]) NewPipe() *Pipe[T] {
	return &Pipe[T]{
		bus:       bus,
		nextEvent: bus.eventQueue.Head(),
	}
}

func (bus *Bus[T]) Listen(listener func(event T)) {
	bus.listeners = append(bus.listeners, listener)
}

func (bus *Bus[T]) Publish(event T) {
	for _, subscriber := range bus.listeners {
		subscriber(event)
	}

	if len(bus.pipes) > 0 {
		bus.eventQueue.Enqueue(event)
	}
}

func (bus *Bus[T]) PublishBatch(events []T) {
	for _, subscriber := range bus.listeners {
		for _, event := range events {
			subscriber(event)
		}
	}

	if len(bus.pipes) > 0 {
		bus.eventQueue.EnqueueBatch(events)
	}
}

func (bus *Bus[T]) dropConsumedQueue() {
	lastAccessibleEvent := bus.pipes[0].nextEvent
	for _, pipe := range bus.pipes[1:] {
		lastAccessibleEvent = min(lastAccessibleEvent, pipe.nextEvent)
	}

	bus.eventQueue.DropUntil(lastAccessibleEvent)
}

func (pipe *Pipe[T]) Iter() iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			event, success := pipe.bus.eventQueue.Peek(pipe.nextEvent)
			if !success {
				break
			}

			pipe.nextEvent++
			if !yield(event) {
				break
			}
		}

		pipe.bus.dropConsumedQueue()
	}
}

func (pipe *Pipe[T]) Pop() (event T, success bool) {
	event, success = pipe.bus.eventQueue.Peek(pipe.nextEvent)

	if success {
		pipe.nextEvent++
		pipe.bus.dropConsumedQueue()
	}

	return
}

func (pipe *Pipe[T]) Close() {
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
