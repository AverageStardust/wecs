package main

import (
	"hash/crc64"
	"reflect"
	"runtime"
	"time"

	"github.com/averagestardust/wecs/internal/common"
	"github.com/averagestardust/wecs/internal/scheduler"
	"github.com/averagestardust/wecs/internal/storage"
)

type System[T any] scheduler.SystemId

type systemState[T any] struct {
	_     struct{} `cbor:",toarray"`
	state *T
}

type systemCallback[T any] func(entity *Access, state *T, delta time.Duration, runtime time.Duration)

var systemCallbacks []reflect.Value

func (system systemState[T]) Run(systemId scheduler.SystemId, store *storage.Store, delta, runtime time.Duration) {
	callback := systemCallbacks[systemId].Interface().(systemCallback[T])

	entity := newAccess(store)
	callback(entity, system.state, delta, runtime)
	entity.Close()
}

func NewSystem[T any](callback systemCallback[T]) System[T] {
	systemId := System[T](len(systemCallbacks))
	systemCallbacks = append(systemCallbacks, reflect.ValueOf(callback))

	return systemId
}

func (system System[T]) Add(schedule *scheduler.Schedule, state *T) {
	schedule.SystemStates[scheduler.SystemId(system)] = systemState[T]{state: state}
}

func (system System[T]) Delete(_schedule *scheduler.Schedule) {
	delete(_schedule.SystemStates, scheduler.SystemId(system))
}

func (system System[T]) Has(schedule *scheduler.Schedule) bool {
	_, hasSystem := schedule.SystemStates[scheduler.SystemId(system)]
	return hasSystem
}

func hashUsedSystemCallbacks(schedules []*scheduler.Schedule) uint64 {
	systemIds := map[scheduler.SystemId]struct{}{}

	for _, schedule := range schedules {
		for systemId := range schedule.SystemStates {
			systemIds[systemId] = struct{}{}
		}
	}

	hash := crc64.New(common.Crc64ISOTable)

	for systemId := range systemIds {
		callback := systemCallbacks[systemId]
		functionPtr := runtime.FuncForPC(callback.Pointer())
		name := functionPtr.Name()
		hash.Write([]byte(name))
	}

	return hash.Sum64()
}
