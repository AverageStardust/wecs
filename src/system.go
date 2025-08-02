package main

import (
	"hash/crc64"
	"reflect"
	"runtime"
	"time"

	"github.com/averagestardust/wecs/internal/common"
	"github.com/averagestardust/wecs/internal/storage"
)

type systemId uint32

type System interface {
	run(store *storage.Store, delta, runtime time.Duration)
	id() systemId
}

type system[T any] struct {
	_        struct{} `cbor:",toarray"`
	systemId systemId
	state    *T
}

type systemCallback[T any] func(access *Access, state *T, delta time.Duration, runtime time.Duration)

var systemCallbacks []reflect.Value
var systemCallbackMap map[reflect.Value]systemId

func NewSystem[T any](schedule *Schedule, state T, callback systemCallback[T]) {
	callbackValue := reflect.ValueOf(callback)
	id, exists := systemCallbackMap[callbackValue]
	if !exists {
		id = systemId(len(systemCallbacks))
		systemCallbacks = append(systemCallbacks, callbackValue)
	}

	system := system[T]{
		systemId: id,
		state:    &state,
	}

	schedule.Systems = append(schedule.Systems, system)
}

func (system system[T]) run(store *storage.Store, delta, runtime time.Duration) {
	callback := systemCallbacks[system.systemId].Interface().(systemCallback[T])

	access := newAccess(store)
	callback(access, system.state, delta, runtime)
	access.Close()
}

func (system system[T]) id() systemId {
	return system.systemId
}

func hashUsedSystemCallbacks(schedules []*Schedule) uint64 {
	systemIds := map[systemId]struct{}{}

	for _, schedule := range schedules {
		for _, system := range schedule.Systems {
			systemIds[system.id()] = struct{}{}
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
