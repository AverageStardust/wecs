package main

import (
	"iter"
	"reflect"
	"unsafe"

	"github.com/averagestardust/wecs/internal/storage"
)

// An integer uniquely identifying a component type.
// Components should store related data on an entity.
// Should be created in a static order during world initialization.
type Component[Data any] storage.PartId

// The next component id to be assigned when a component is created.
// Increments from zero.
var nextComponent storage.PartId = 0

// Create a new component type from a data type.
// The data type should store related data, probably with a struct.
// Should be used in a static order during world initialization.
func NewComponent[Data any]() Component[Data] {
	var data Data
	typ := reflect.TypeOf(data)

	storage.NewPartType(nextComponent, typ)
	component := Component[Data](nextComponent)

	nextComponent++
	return component
}

// Remove a component from an entity.
func (component Component[Data]) Delete(access *Access, entity Entity) (success bool) {
	return access.store.DeletePart(storage.EntityId(entity), component)
}

// Get the data of a component from an entity.
func (component Component[Data]) Get(access *Access, entity Entity) (data *Data) {
	bytes := access.store.GetComponent(storage.EntityId(entity), storage.PartId(component))
	if bytes == nil {
		return nil
	}

	return (*Data)(unsafe.Pointer(&bytes[0]))
}

// Check if a entity has a component.
func (component Component[Data]) Has(access *Access, entity Entity) (success bool) {
	return access.store.HasPart(storage.EntityId(entity), component)
}

// Add a component with empty data to an entity.
func (component Component[Data]) Add(access *Access, entity Entity) (success bool) {
	return access.store.AddPart(storage.EntityId(entity), component)
}

// Return an iterator of data from one component type from all entities that match a filter.
func (component Component[Data]) Query(access *Access, filter Filter) iter.Seq[*Data] {
	return func(yield func(*Data) bool) {
		for page := range filter.filter(access.store) {
			for bytes := range page.GetComponentIter(storage.PartId(component)) {
				data := (*Data)(unsafe.Pointer(&bytes))
				if !yield(data) {
					return
				}
			}
		}
	}
}

// Get the part id.
func (component Component[Data]) PartId() storage.PartId {
	return storage.PartId(component)
}
