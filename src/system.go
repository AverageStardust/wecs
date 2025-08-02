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

type System[T any] struct {
	_     struct{} `cbor:",toarray"`
	id    scheduler.SystemId
	state *T
}

type systemCallback[T any] func(access *Access, state *T, delta time.Duration, runtime time.Duration)

var systemCallbacks []reflect.Value

func NewSystem[T any](state T, callback systemCallback[T]) System[T] {
	systemId := scheduler.SystemId(len(systemCallbacks))
	systemCallbacks = append(systemCallbacks, reflect.ValueOf(callback))

	return System[T]{
		id:    systemId,
		state: &state,
	}
}

func (system System[T]) Run(store *storage.Store, delta, runtime time.Duration) {
	callback := systemCallbacks[system.id].Interface().(systemCallback[T])

	access := newAccess(store)
	callback(access, system.state, delta, runtime)
	access.Close()
}

func (system System[T]) Id() scheduler.SystemId {
	return system.id
}

func hashUsedSystemCallbacks(schedules []*scheduler.Schedule) uint64 {
	systemIds := map[scheduler.SystemId]struct{}{}

	for _, schedule := range schedules {
		for _, system := range schedule.Systems {
			systemIds[system.Id()] = struct{}{}
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
