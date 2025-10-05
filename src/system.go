package main

import (
	"time"
)

// An interface for systems that finds entities and manipulates their components.
type System interface {
	Run(world *World, delta time.Duration)
	State() any
	Runtime() time.Duration
	SetRuntime(runtime time.Duration)
}

// A system with state that finds entities and manipulates their components.
type system[T any] struct {
	state    *T
	callback systemCallback[T]
	runtime  time.Duration
}

// A function callback that runs a system.
type systemCallback[T any] func(world *World, state *T, delta time.Duration, runtime time.Duration)

func NewSystem[T any](state T, callback systemCallback[T]) System {
	return &system[T]{
		state:    &state,
		callback: callback,
		runtime:  time.Duration(0),
	}
}

// Run a system using it's state.
func (system *system[T]) Run(world *World, delta time.Duration) {
	system.callback(world, system.state, delta, system.runtime)
	world.EmptyDeleteQueue()
	system.runtime += delta
}

func (system *system[T]) State() any {
	return system.state
}

func (system *system[T]) Runtime() time.Duration {
	return system.runtime
}

func (system *system[T]) SetRuntime(runtime time.Duration) {
	system.runtime = runtime
}
