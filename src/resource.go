package main

import (
<<<<<<< HEAD
=======
	"log"
>>>>>>> 0809119 (simplify resource names)
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

<<<<<<< HEAD
func NewResource[T any]() Resource[T] {
	var t T
	var typ = reflect.TypeOf(t)
	var name string

=======
// a set of resource types
var resourceTypeSet = map[string]struct{}{}

func NewResource[T any]() (resourceType Resource[T]) {
	var t T
	var typ = reflect.TypeOf(t)

	var name string
>>>>>>> 0809119 (simplify resource names)
	if typ.PkgPath() != "" {
		name = typ.PkgPath() + "." + typ.Name()
	} else {
		name = typ.String()
	}

<<<<<<< HEAD
=======
	_, duplicate := resourceTypeSet[name]
	if duplicate {
		log.Println("WARNING: resources must have unique, can't distinguish type resource of type \"" + name + "\"")
	}

	resourceTypeSet[typ] = struct{}{}

>>>>>>> 0809119 (simplify resource names)
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
