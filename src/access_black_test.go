package main_test

import (
	"testing"

	wecs "github.com/averagestardust/wecs"
	"github.com/stretchr/testify/assert"
)

func TestAccessQuery(t *testing.T) {
	world := wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()
	Vector := wecs.NewComponent[struct {
		x float32
		y float32
	}]()

	world.New(Integer)
	vectorEntity := world.New(Vector)
	world.New(Integer, Vector)

	for entity := range world.Query(wecs.NewFilter().Exactly(Vector)) {
		assert.Equal(t, entity, vectorEntity)
	}
}

func TestAccessAlive(t *testing.T) {
	world := wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()

	fake := wecs.Entity(2304)
	assert.False(t, world.Alive(fake))

	entity := world.New(Integer)
	assert.True(t, world.Alive(entity))

	world.Delete(entity)
	assert.False(t, world.Alive(entity))

	world.DeleteImmediately(entity)
	assert.False(t, world.Alive(entity))
}

func TestAccessExists(t *testing.T) {
	world := wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()

	fake := wecs.Entity(2304)
	assert.False(t, world.Exists(fake))

	entity := world.New(Integer)
	assert.True(t, world.Exists(entity))

	world.Delete(entity)
	assert.True(t, world.Exists(entity))

	world.DeleteImmediately(entity)
	assert.False(t, world.Exists(entity))
}

func TestAccessEmptyDeleteQueue(t *testing.T) {
	world := wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()

	entity := world.New(Integer)
	assert.True(t, world.Alive(entity))
	assert.True(t, world.Exists(entity))

	world.Delete(entity)
	assert.False(t, world.Alive(entity))
	assert.True(t, world.Exists(entity))

	world.EmptyDeleteQueue()
	assert.False(t, world.Alive(entity))
	assert.False(t, world.Exists(entity))
}

func TestAccessDelete(t *testing.T) {
	world := wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()

	entity := world.New(Integer)
	assert.True(t, world.Alive(entity))
	assert.True(t, world.Exists(entity))

	world.Delete(entity)
	assert.False(t, world.Alive(entity))
	assert.True(t, world.Exists(entity))

	world.EmptyDeleteQueue()
	assert.False(t, world.Alive(entity))
	assert.False(t, world.Exists(entity))
}

func TestAccessDeleteImmediately(t *testing.T) {
	world := wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()

	entity := world.New(Integer)
	assert.True(t, world.Alive(entity))
	assert.True(t, world.Exists(entity))

	world.DeleteImmediately(entity)
	assert.False(t, world.Alive(entity))
	assert.False(t, world.Exists(entity))
}

func TestAccessNew(t *testing.T) {
	world := wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()
	Vector := wecs.NewComponent[struct {
		x float32
		y float32
	}]()

	entityA := world.New(Integer)
	entityB := world.New(Integer)
	entityC := world.New(Integer, Vector)
	entityD := world.New(Vector)

	assert.True(t, Integer.Has(world, entityA))
	assert.False(t, Vector.Has(world, entityA))

	assert.True(t, Integer.Has(world, entityB))
	assert.False(t, Vector.Has(world, entityB))

	assert.True(t, Integer.Has(world, entityC))
	assert.True(t, Vector.Has(world, entityC))

	assert.False(t, Integer.Has(world, entityD))
	assert.True(t, Vector.Has(world, entityD))

	assert.NotEqual(t, entityA, entityB)
	assert.NotEqual(t, entityA, entityC)
	assert.NotEqual(t, entityA, entityD)
	assert.NotEqual(t, entityB, entityC)
	assert.NotEqual(t, entityB, entityD)
	assert.NotEqual(t, entityC, entityD)
}

func TestAccessNewBatch(t *testing.T) {
	world := wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()
	Vector := wecs.NewComponent[struct {
		x float32
		y float32
	}]()

	count := 0
	for entity := range world.NewBatch(25, Integer) {
		assert.True(t, Integer.Has(world, entity))
		assert.False(t, Vector.Has(world, entity))
		count++
	}

	assert.Equal(t, 25, count)
}
