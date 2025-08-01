package main_test

import (
	"testing"

	wecs "github.com/averagestardust/wecs"
	"github.com/stretchr/testify/assert"
)

func TestWorldGetStorage(t *testing.T) {
	world := wecs.NewWorld()
	Integer := wecs.NewComponent[int]()
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

	for range 16 {
		go func() {
			world.GetAccess(func(access *wecs.Access) {
				access.New(Integer)
			})
		}()
	}
}
