package main

import (
	"iter"
	"unsafe"

	"github.com/averagestardust/wecs/internal/storage"
)

// A pair of components often used together
type Pair[T any, U any] [2]storage.PartId

// Create a pair of components to query for data together
func NewPair[T any, U any](a Component[T], b Component[U]) Pair[T, U] {
	return Pair[T, U]{storage.PartId(a), storage.PartId(b)}
}

// Return an iterator of data from two component types from all entities that match a filter.
func (pair Pair[T, U]) Query(world *World, filter Filter) iter.Seq2[*T, *U] {
	return func(yield func(*T, *U) bool) {
		for page := range filter.filter(world.store) {
			nextA, stopA := iter.Pull(page.GetComponentIter(pair[0]))
			nextB, stopB := iter.Pull(page.GetComponentIter(pair[1]))
			defer stopA()
			defer stopB()

			for {
				aBytes, done := nextA()
				bBytes, _ := nextB()
				aComponent := (*T)(unsafe.Pointer(&aBytes))
				bComponent := (*U)(unsafe.Pointer(&bBytes))

				// yield both components
				if !yield(aComponent, bComponent) {
					return
				}
				if done {
					return
				}
			}
		}

	}
}
