package main

import (
	"errors"
	"io"
	"time"

	"github.com/averagestardust/wecs/internal/storage"
	"github.com/fxamacker/cbor/v2"
)

type systemSave struct {
	_       struct{} `cbor:",toarray"`
	State   any
	Runtime time.Duration
}

type worldSave struct {
	_           struct{} `cbor:",toarray"`
	Store       *storage.Store
	DeleteQueue map[Entity]struct{}
	PartHash    uint64
}

var ErrIncompatibleParts = errors.New("can't deserialize because existing parts don't match save")

func SerializeSystem(system System, writer io.Writer) (err error) {
	return Serialize(systemSave{
		State:   system.State(),
		Runtime: system.Runtime(),
	}, writer)
}

func DeserializeSystem[T any](callback systemCallback[T], reader io.Reader) (system System, err error) {
	save, err := Deserialize[*systemSave](reader)
	if err != nil {
		return nil, err
	}

	system = NewSystem(save.State.(T), callback)
	system.SetRuntime(save.Runtime)

	return system, err
}

func SerializeWorld(world *World, writer io.Writer) (err error) {
	partHash := storage.HashUsedParts(world.store)

	return Serialize(worldSave{
		Store:       world.store,
		DeleteQueue: world.deleteQueue,
		PartHash:    partHash,
	}, writer)
}

func DeserializeWorld(reader io.Reader) (world *World, err error) {
	save, err := Deserialize[*worldSave](reader)
	if err != nil {
		return nil, err
	}

	world = NewWorld()
	world.store = save.Store
	world.deleteQueue = save.DeleteQueue
	world.EmptyDeleteQueue()

	if save.PartHash != storage.HashUsedParts(world.store) {
		return nil, ErrIncompatibleParts
	}

	return world, err
}

func Serialize(object any, writer io.Writer) (err error) {
	return cbor.NewEncoder(writer).Encode(object)
}

func Deserialize[T any](reader io.Reader) (object T, err error) {
	err = cbor.NewDecoder(reader).Decode(object)
	return
}
