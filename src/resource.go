package main

import (
	"log"
	"reflect"

	"github.com/averagestardust/wecs/internal/storage"
)

type GenericResource interface {
	Create() any
	Name() string
}

type Resource[T any] struct {
	name string
}

// a set of resource types
var resourceTypeSet = map[string]struct{}{}

func NewResource[T any]() (resourceType Resource[T]) {
	var t T
	var typ = reflect.TypeOf(t)

	var name string
	if typ.PkgPath() != "" {
		name = typ.PkgPath() + "." + typ.Name()
	} else {
		name = typ.String()
	}

	_, duplicate := resourceTypeSet[name]
	if duplicate {
		log.Println("WARNING: resources must be unique, can't distinguish resource of type \"" + name + "\"")
	}

	resourceTypeSet[name] = struct{}{}

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
