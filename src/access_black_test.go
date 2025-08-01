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

	world.GetAccess(func(access *wecs.Access) {
		access.New(Integer)
		vectorEntity := access.New(Vector)
		access.New(Integer, Vector)

		for entity := range access.Query(wecs.NewFilter().Exactly(Vector)) {
			assert.Equal(t, entity, vectorEntity)
		}
	})
}

func TestAccessAlive(t *testing.T) {
	world := wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()

	world.GetAccess(func(access *wecs.Access) {
		fake := wecs.Entity(2304)
		assert.False(t, access.Alive(fake))

		entity := access.New(Integer)
		assert.True(t, access.Alive(entity))

		access.Delete(entity)
		assert.False(t, access.Alive(entity))

		access.DeleteImmediately(entity)
		assert.False(t, access.Alive(entity))
	})
}

func TestAccessExists(t *testing.T) {
	world := *wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()

	world.GetAccess(func(access *wecs.Access) {
		fake := wecs.Entity(2304)
		assert.False(t, access.Exists(fake))

		entity := access.New(Integer)
		assert.True(t, access.Exists(entity))

		access.Delete(entity)
		assert.True(t, access.Exists(entity))

		access.DeleteImmediately(entity)
		assert.False(t, access.Exists(entity))
	})
}

func TestAccessEmptyDeleteQueue(t *testing.T) {
	world := *wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()

	world.GetAccess(func(access *wecs.Access) {
		entity := access.New(Integer)
		assert.True(t, access.Alive(entity))
		assert.True(t, access.Exists(entity))

		access.Delete(entity)
		assert.False(t, access.Alive(entity))
		assert.True(t, access.Exists(entity))

		access.EmptyDeleteQueue()
		assert.False(t, access.Alive(entity))
		assert.False(t, access.Exists(entity))
	})
}

func TestAccessDelete(t *testing.T) {
	world := *wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()

	var entity wecs.Entity

	world.GetAccess(func(access *wecs.Access) {
		entity = access.New(Integer)
		assert.True(t, access.Alive(entity))
		assert.True(t, access.Exists(entity))

		access.Delete(entity)
		assert.False(t, access.Alive(entity))
		assert.True(t, access.Exists(entity))
	})

	world.GetAccess(func(access *wecs.Access) {
		assert.False(t, access.Alive(entity))
		assert.False(t, access.Exists(entity))
	})
}

func TestAccessDeleteImmediately(t *testing.T) {
	world := *wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()

	var entity wecs.Entity

	world.GetAccess(func(access *wecs.Access) {
		entity = access.New(Integer)
		assert.True(t, access.Alive(entity))
		assert.True(t, access.Exists(entity))

		access.DeleteImmediately(entity)
		assert.False(t, access.Alive(entity))
		assert.False(t, access.Exists(entity))
	})

	world.GetAccess(func(access *wecs.Access) {
		assert.False(t, access.Alive(entity))
		assert.False(t, access.Exists(entity))
	})
}

func TestAccessNew(t *testing.T) {
	world := *wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()
	Vector := wecs.NewComponent[struct {
		x float32
		y float32
	}]()

	world.GetAccess(func(access *wecs.Access) {
		entityA := access.New(Integer)
		entityB := access.New(Integer)
		entityC := access.New(Integer, Vector)
		entityD := access.New(Vector)

		assert.True(t, Integer.Has(access, entityA))
		assert.False(t, Vector.Has(access, entityA))

		assert.True(t, Integer.Has(access, entityB))
		assert.False(t, Vector.Has(access, entityB))

		assert.True(t, Integer.Has(access, entityC))
		assert.True(t, Vector.Has(access, entityC))

		assert.False(t, Integer.Has(access, entityD))
		assert.True(t, Vector.Has(access, entityD))

		assert.NotEqual(t, entityA, entityB)
		assert.NotEqual(t, entityA, entityC)
		assert.NotEqual(t, entityA, entityD)
		assert.NotEqual(t, entityB, entityC)
		assert.NotEqual(t, entityB, entityD)
		assert.NotEqual(t, entityC, entityD)
	})
}

func TestAccessNewBatch(t *testing.T) {
	world := *wecs.NewWorld()
	Integer := wecs.NewComponent[uint32]()
	Vector := wecs.NewComponent[struct {
		x float32
		y float32
	}]()

	world.GetAccess(func(access *wecs.Access) {
		count := 0
		for entity := range access.NewBatch(25, Integer) {
			assert.True(t, Integer.Has(access, entity))
			assert.False(t, Vector.Has(access, entity))
			count++
		}

		assert.Equal(t, 25, count)
	})
}
