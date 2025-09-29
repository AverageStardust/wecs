package main

import (
	"hash/crc64"
	"reflect"
	"runtime"
	"time"

	"github.com/averagestardust/wecs/internal/common"
	"github.com/averagestardust/wecs/internal/storage"
)

// An integer uniquely identifying a system callback.
type systemId uint32

// An interface for systems that finds entities and manipulates their components.
type System interface {
	run(store *storage.Store, delta, runtime time.Duration)
	id() systemId
}

// A system with state that finds entities and manipulates their components.
type system[T any] struct {
	_        struct{} `cbor:",toarray"`
	systemId systemId
	state    *T
}

// A function callback that runs a system.
type systemCallback[T any] func(access *Access, state *T, delta time.Duration, runtime time.Duration)

// A list of callbacks for systems
var systemCallbacks []reflect.Value

// A dictionary to find the system ids of callbacks
var systemCallbackMap map[reflect.Value]systemId

// Add a new system on a schedule.
// Should be used in a static order during world initialization.
func NewSystem[T any](schedule Schedule, state T, callback systemCallback[T]) {
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

	schedule.appendSystem(system)
}

// Run a system using it's state.
func (system system[T]) run(store *storage.Store, delta, runtime time.Duration) {
	callback := systemCallbacks[system.systemId].Interface().(systemCallback[T])

	access := newAccess(store)
	callback(access, system.state, delta, runtime)
	access.Close()
}

// Get the id of a system.
func (system system[T]) id() systemId {
	return system.systemId
}

// Hash all the system callbacks used on a set of schedules.
func hashUsedSystemCallbacks(schedules []*schedule) uint64 {
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
