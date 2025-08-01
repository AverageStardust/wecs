package main

import (
	"iter"

	"github.com/averagestardust/wecs/internal/storage"
)

type Access struct {
	store       *storage.Store
	deleteQueue map[Entity]struct{}
}

func newAccess(store *storage.Store) *Access {
	store.Mutex.Lock()
	return &Access{
		store:       store,
		deleteQueue: map[Entity]struct{}{},
	}
}

func (access *Access) Close() {
	access.EmptyDeleteQueue()
	access.store.Mutex.Unlock()
	access.store = nil
}

func (access *Access) Alive(entity Entity) bool {
	_, deleteIsQueued := access.deleteQueue[entity]
	return !deleteIsQueued && access.Exists(entity)
}

func (access *Access) Exists(entity Entity) bool {
	_, exists := access.store.Entries[storage.EntityId(entity)]
	return exists
}

func (access *Access) Query(filter Filter) iter.Seq[Entity] {
	return func(yield func(Entity) bool) {
		for page := range filter.Filter(access.store) {
			for _, entity := range page.Access {
				if !yield(Entity(entity)) {
					return
				}
			}
		}
	}
}

func (access *Access) EmptyDeleteQueue() {
	for entity := range access.deleteQueue {
		access.store.Delete(storage.EntityId(entity))
		delete(access.deleteQueue, entity)
	}
}

func (access *Access) Delete(entity Entity) {
	access.deleteQueue[entity] = struct{}{}
}

func (access *Access) DeleteImmediately(entity Entity) {
	access.store.Delete(storage.EntityId(entity))
}

func (access *Access) New(parts ...storage.Part) Entity {
	archetype := access.store.NewArchetype(parts)
	entity := access.store.Grow(archetype, 1)

	return Entity(entity)
}

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
