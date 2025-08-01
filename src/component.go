package main

import (
	"iter"
	"reflect"
	"unsafe"

	"github.com/averagestardust/wecs/internal/storage"
)

type Component[T any] storage.PartId

var nextComponent storage.PartId = 0

func NewComponent[T any]() Component[T] {
	var t T
	typ := reflect.TypeOf(t)

	storage.NewPartType(nextComponent, typ)
	component := Component[T](nextComponent)

	nextComponent++
	return component
}

func (component Component[T]) Delete(access *Access, entity Entity) bool {
	return access.store.DeletePart(storage.EntityId(entity), component)
}

func (component Component[T]) Get(access *Access, entity Entity) *T {
	bytes := access.store.GetComponent(storage.EntityId(entity), storage.PartId(component))
	if bytes == nil {
		return nil
	}

	return (*T)(unsafe.Pointer(&bytes[0]))
}

func (component Component[T]) Has(access *Access, entity Entity) bool {
	return access.store.HasPart(storage.EntityId(entity), component)
}

func (component Component[T]) Add(access *Access, entity Entity) bool {
	return access.store.AddPart(storage.EntityId(entity), component)
}

func (component Component[T]) Query(access *Access, filter Filter) iter.Seq[*T] {
	return func(yield func(*T) bool) {
		for page := range filter.Filter(access.store) {
			for bytes := range page.GetComponentIter(storage.PartId(component)) {
				component := (*T)(unsafe.Pointer(&bytes))
				if !yield(component) {
					return
				}
			}
		}
	}
}

func (component Component[T]) PartId() storage.PartId {
	return storage.PartId(component)
}
