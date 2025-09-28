package main

import (
	"iter"

	"github.com/averagestardust/wecs/internal/storage"
)

// A thread-safe access to the contents of a World.
// It also allows entities deletions to be queued until the access is closed.
type Access struct {
	store       *storage.Store
	deleteQueue map[Entity]struct{}
}

// Lock a thread-safe access to the contents of a World.
func newAccess(store *storage.Store) *Access {
	store.Mutex.Lock()
	return &Access{
		store:       store,
		deleteQueue: map[Entity]struct{}{},
	}
}

// Release access to the contents of a World.
func (access *Access) Close() {
	access.EmptyDeleteQueue()
	access.store.Mutex.Unlock()
	access.store = nil
}

// Check if an entity is exists and hasn't been queued for deletion.
func (access *Access) Alive(entity Entity) bool {
	_, deleteIsQueued := access.deleteQueue[entity]
	return !deleteIsQueued && access.Exists(entity)
}

// Check if an entity exists.
func (access *Access) Exists(entity Entity) bool {
	_, exists := access.store.Entries[storage.EntityId(entity)]
	return exists
}

// Return an iterator of all the entities that match the filter.
func (access *Access) Query(filter Filter) iter.Seq[Entity] {
	return func(yield func(Entity) bool) {
		for page := range filter.filter(access.store) {
			for _, entity := range page.Entities {
				if !yield(Entity(entity)) {
					return
				}
			}
		}
	}
}

// Immediately delete all entities that have been queued for deletions.
func (access *Access) EmptyDeleteQueue() {
	for entity := range access.deleteQueue {
		access.store.Delete(storage.EntityId(entity))
		delete(access.deleteQueue, entity)
	}
}

// Queue an entity for deletion after the access is closed.
func (access *Access) Delete(entity Entity) {
	access.deleteQueue[entity] = struct{}{}
}

// Immediately delete an entity, without queuing it.
func (access *Access) DeleteImmediately(entity Entity) {
	access.store.Delete(storage.EntityId(entity))
}

// Create a new entity out of an arbitrary list of components/tags.
func (access *Access) New(parts ...storage.Part) Entity {
	archetype := access.store.NewArchetype(parts)
	entity := access.store.Grow(archetype, 1)

	return Entity(entity)
}

// Create multiple identical new entities out of an arbitrary list of components/tags.
// Returns a iterator of the new entities.
func (access *Access) NewBatch(count int, parts ...storage.Part) iter.Seq[Entity] {
	archetype := access.store.NewArchetype(parts)
	firstEntity := access.store.Grow(archetype, count)

	return func(yield func(Entity) bool) {
		for i := range count {
			if !yield(Entity(firstEntity + storage.EntityId(i))) {
				return
			}
		}
	}
}
