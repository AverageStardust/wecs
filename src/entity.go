package main

import (
	"iter"

	"github.com/averagestardust/wecs/internal/storage"
)

// An integer uniquely identifying an entity over the life of the program
type Entity storage.EntityId

// Check if an entity exists.
func (world *World) Exists(entity Entity) bool {
	_, exists := world.store.Entries[storage.EntityId(entity)]
	return exists
}

// Return an iterator of all the entities that match the filter.
func (world *World) Query(filter Filter) iter.Seq[Entity] {
	return func(yield func(Entity) bool) {
		for page := range filter.filter(world.store) {
			for _, entity := range page.Entities {
				if !yield(Entity(entity)) {
					return
				}
			}
		}
	}
}

// Immediately delete an entity, without queuing it.
func (world *World) Delete(entity Entity) {
	world.store.Delete(storage.EntityId(entity))
}

// Create a new entity out of an arbitrary list of components/tags.
func (world *World) New(parts ...storage.Part) Entity {
	archetype := world.store.NewArchetype(parts)
	entity := world.store.Grow(archetype, 1)

	return Entity(entity)
}

// Create multiple identical new entities out of an arbitrary list of components/tags.
// Returns a iterator of the new entities.
func (world *World) NewBatch(count int, parts ...storage.Part) iter.Seq[Entity] {
	archetype := world.store.NewArchetype(parts)
	firstEntity := world.store.Grow(archetype, count)

	return func(yield func(Entity) bool) {
		for i := range count {
			if !yield(Entity(firstEntity + storage.EntityId(i))) {
				return
			}
		}
	}
}
