package main

import (
	"github.com/averagestardust/wecs/internal/storage"
)

type World struct {
	store       *storage.Store
	deleteQueue map[Entity]struct{}
}

// Create a new world.
func NewWorld() *World {
	return &World{
		store:       storage.NewStore(),
		deleteQueue: map[Entity]struct{}{},
	}
}

// Check if an entity is exists and hasn't been queued for deletion.
func (world *World) Alive(entity Entity) bool {
	_, deleteIsQueued := world.deleteQueue[entity]
	return !deleteIsQueued && world.Exists(entity)
}

// Queue an entity for deletion after the access is closed.
func (world *World) QueueDelete(entity Entity) {
	world.deleteQueue[entity] = struct{}{}
}

// Immediately delete all entities that have been queued for deletions.
func (world *World) EmptyDeleteQueue() {
	for entity := range world.deleteQueue {
		world.store.Delete(storage.EntityId(entity))
		delete(world.deleteQueue, entity)
	}
}
