package main

import (
	"iter"
	"unsafe"

	"github.com/averagestardust/wecs/internal/storage"
)

type Pair[T any, U any] [2]storage.PartId

func NewPair[T any, U any](a Component[T], b Component[U]) Pair[T, U] {
	return Pair[T, U]{storage.PartId(a), storage.PartId(b)}
}

func (pair Pair[T, U]) Query(access *Access, filter Filter) iter.Seq2[*T, *U] {
	return func(yield func(*T, *U) bool) {
		for page := range filter.filter(access.store) {
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
