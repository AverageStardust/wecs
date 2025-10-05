package main

import (
	"errors"
	"io"
	"iter"
	"time"

	"github.com/averagestardust/wecs/internal/storage"
	"github.com/fxamacker/cbor/v2"
)

type worldSave struct {
	_            struct{} `cbor:",toarray"`
	Store        *storage.Store
	Scheduler    *scheduler
	DeleteQueue  map[Entity]struct{}
	CallbackHash uint64
	PartHash     uint64
	ResourceHash uint64
	Version      int
}

type World struct {
	store       *storage.Store
	scheduler   *scheduler
	deleteQueue map[Entity]struct{}
}

var ErrIncompatibleCallbacks = errors.New("can't deserialize because existing system callbacks don't match save")
var ErrIncompatibleParts = errors.New("can't deserialize because existing parts don't match save")
var ErrIncompatibleResources = errors.New("can't deserialize because existing resources don't match save")
var ErrIncompatibleVersion = errors.New("can't deserialize because current version is older than save")

// Create a new world.
func NewWorld() *World {
	return &World{
		store:       storage.NewStore(),
		scheduler:   newScheduler(),
		deleteQueue: map[Entity]struct{}{},
	}
}

// Stop all systems from running on a system.
func (world *World) StopSchedules() {
	world.scheduler.stop()
}

// Start running all systems on a system.
func (world *World) RunSchedules() {
	world.scheduler.run(world)
}

// Step a particular schedule on a world.
func (world *World) StepSchedule(schedule Schedule) {
	schedule.run(world, time.Now())
}

// Create a new schedule that must be manual stepped using StepSchedule().
func (world *World) NewManualSchedule(frequency float64) Schedule {
	return world.scheduler.newManualSchedule()
}

// Create a new schedule that will automatically step it's systems itself.
func (world *World) NewSchedule(frequency float64) Schedule {
	return world.scheduler.newSchedule(frequency, frequency)
}

// Create a new schedule that will automatically step at a range of frequencies.
func (world *World) NewVariableSchedule(maxFrequency float64, minFrequency float64) Schedule {
	return world.scheduler.newSchedule(maxFrequency, minFrequency)
}

// Check if an entity is exists and hasn't been queued for deletion.
func (world *World) Alive(entity Entity) bool {
	_, deleteIsQueued := world.deleteQueue[entity]
	return !deleteIsQueued && world.Exists(entity)
}

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

// Immediately delete all entities that have been queued for deletions.
func (world *World) EmptyDeleteQueue() {
	for entity := range world.deleteQueue {
		world.store.Delete(storage.EntityId(entity))
		delete(world.deleteQueue, entity)
	}
}

// Queue an entity for deletion after the access is closed.
func (world *World) Delete(entity Entity) {
	world.deleteQueue[entity] = struct{}{}
}

// Immediately delete an entity, without queuing it.
func (world *World) DeleteImmediately(entity Entity) {
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

// Deserialize the content of the world, overwriting whatever entities are there.
// Must already have components types, tag types, resource types and systems setup.
func (world *World) Deserialize(version int, reader io.Reader) error {
	save := &worldSave{}
	err := cbor.NewDecoder(reader).Decode(save)
	if err != nil {
		return err
	}

	if save.CallbackHash != hashUsedSystemCallbacks(save.Scheduler.Schedules) {
		return ErrIncompatibleCallbacks
	} else if save.PartHash != storage.HashUsedParts(world.store) {
		return ErrIncompatibleParts
	} else if save.ResourceHash != hashUsedResources(save.Store) {
		return ErrIncompatibleResources
	} else if save.Version > version {
		return ErrIncompatibleVersion
	}

	stopped := world.scheduler.stop()

	world.store = storage.NewStore()
	world.store = save.Store
	world.scheduler = newScheduler()
	world.scheduler = save.Scheduler

	// reset tickers
	for _, schedule := range world.scheduler.Schedules {
		schedule.resetTicker()
	}

	if stopped {
		world.scheduler.run(world)
	}

	world.EmptyDeleteQueue()

	return nil
}

// Serialize the content of the world.
func (world *World) Serialize(version int, writer io.Writer) error {
	stopped := world.scheduler.stop()

	save := worldSave{
		Store:        world.store,
		Scheduler:    world.scheduler,
		DeleteQueue:  world.deleteQueue,
		CallbackHash: hashUsedSystemCallbacks(world.scheduler.Schedules),
		PartHash:     storage.HashUsedParts(world.store),
		ResourceHash: hashUsedResources(world.store),
		Version:      version,
	}

	err := cbor.NewEncoder(writer).Encode(save)

	if stopped {
		world.scheduler.run(world)
	}

	return err
}
