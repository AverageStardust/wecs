package main_test

import (
	"testing"
	"time"

	wecs "github.com/averagestardust/wecs"
	"github.com/averagestardust/wecs/internal/scheduler"
	"github.com/averagestardust/wecs/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestSystem(t *testing.T) {
	scheduler := scheduler.NewScheduler()
	schedule := scheduler.NewSchedule(50, 50)

	iterations := 0

	counterSystem := wecs.NewSystem(struct{}{}, func(_ *wecs.Access, state *struct{}, delta time.Duration, _ time.Duration) {
		assert.InDelta(t, time.Millisecond*20, delta, float64(time.Nanosecond))
		iterations++
	})

	schedule.Add(counterSystem)

	store := storage.NewStore()
	go scheduler.RunSystems(store)

	time.Sleep(time.Millisecond * 200)

	scheduler.StopSystems()

	assert.InDelta(t, 10, iterations, 2)
	assert.InDelta(t, time.Millisecond*200, schedule.RunTime, float64(time.Millisecond*40))
	assert.WithinRange(t, schedule.LastTime, time.Now().Add(-time.Millisecond*50), time.Now())
}
