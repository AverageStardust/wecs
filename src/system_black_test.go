package main_test

import (
	"testing"
	"time"

	wecs "github.com/averagestardust/wecs"
	"github.com/stretchr/testify/assert"
)

func TestSystem(t *testing.T) {
	world := wecs.NewWorld()
	schedule := world.NewSchedule(50)

	iterations := 0

	wecs.NewSystem(schedule, struct{}{}, func(_ *wecs.Access, state *struct{}, delta time.Duration, runtime time.Duration) {
		assert.InDelta(t, time.Millisecond*20, delta, float64(time.Nanosecond))
		assert.InDelta(t, time.Millisecond*time.Duration(20*iterations), runtime, float64(time.Nanosecond))
		iterations++
	})

	go world.RunSchedules()
	time.Sleep(time.Millisecond * 200)

	world.StopSchedules()

	assert.InDelta(t, 10, iterations, 2)
}
