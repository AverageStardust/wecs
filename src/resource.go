package main

import (
	"github.com/averagestardust/wecs/internal/storage"
)

type GenericResource interface {
	Create() any
	Name() string
}

type Resource[T any] struct {
	name string
}

func NewResource[T any](name string, version int) Resource[T] {
	return Resource[T]{
		name,
	}
}

// Get a resource from the world.
func (resourceType Resource[T]) Get(store *storage.Store) *T {
	resource, exists := store.Resources[resourceType.name]
	if !exists {
		return nil
	}

	return resource.(*T)
}

// Check if the world has a resource.
func (resourceType Resource[T]) Has(store *storage.Store) bool {
	_, exists := store.Resources[resourceType.name]
	return exists
}

func (resourceType Resource[T]) Create() any {
	var t T
	return t
}

func (resourceType Resource[T]) Name() string {
	return resourceType.name
}
