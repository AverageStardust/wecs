package main

import (
	"errors"
	"io"
	"time"

	"github.com/averagestardust/wecs/internal/storage"
	"github.com/fxamacker/cbor/v2"
)

type worldSave struct {
	_            struct{} `cbor:",toarray"`
	Store        *storage.Store
	Scheduler    *scheduler
	CallbackHash uint64
	PartHash     uint64
	ResourceHash uint64
	Version      int
}

type World struct {
	store     *storage.Store
	scheduler *scheduler
}

var ErrIncompatibleCallbacks = errors.New("can't deserialize because existing system callbacks don't match save")
var ErrIncompatibleParts = errors.New("can't deserialize because existing parts don't match save")
var ErrIncompatibleResources = errors.New("can't deserialize because existing resources don't match save")
var ErrIncompatibleVersion = errors.New("can't deserialize because current version is older than save")

func NewWorld() *World {
	return &World{
		store:     storage.NewStore(),
		scheduler: newScheduler(),
	}
}

func (world *World) StopSchedules() {
	world.scheduler.stop()
}

func (world *World) RunSchedules() {
	world.scheduler.run(world.store)
}

func (world *World) StepSchedule(schedule Schedule) {
	schedule.run(world.store, time.Now())
}

func (world *World) NewManuelSchedule(frequency float64) Schedule {
	return world.scheduler.newManuelSchedule()
}

func (world *World) NewSchedule(frequency float64) Schedule {
	return world.scheduler.newSchedule(frequency, frequency)
}

func (world *World) NewVariableSchedule(maxFrequency float64, minFrequency float64) Schedule {
	return world.scheduler.newSchedule(maxFrequency, minFrequency)
}

func (world *World) GetAccess(callback func(access *Access)) {
	access := newAccess(world.store)
	callback(access)
	access.Close()
}

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
		world.scheduler.run(world.store)
	}

	return nil
}

func (world *World) Serialize(version int, writer io.Writer) error {
	stopped := world.scheduler.stop()

	world.store.Mutex.Lock()
	save := worldSave{
		Store:        world.store,
		Scheduler:    world.scheduler,
		CallbackHash: hashUsedSystemCallbacks(world.scheduler.Schedules),
		PartHash:     storage.HashUsedParts(world.store),
		ResourceHash: hashUsedResources(world.store),
		Version:      version,
	}

	err := cbor.NewEncoder(writer).Encode(save)

	world.store.Mutex.Unlock()

	if stopped {
		world.scheduler.run(world.store)
	}

	return err
}
