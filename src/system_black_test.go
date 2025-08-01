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
	schedule50hz := scheduler.NewSchedule(50)

	counterSystem := wecs.NewSystem(func(_ *wecs.Access, state *int, delta time.Duration, _ time.Duration) {
		assert.InDelta(t, time.Millisecond*20, delta, float64(time.Nanosecond))
		*state++
	})

	var counterState = 0
	counterSystem.Add(schedule50hz, &counterState)

	store := storage.NewStore()
	go scheduler.RunSystems(store)

	time.Sleep(time.Millisecond * 200)

	scheduler.StopSystems()

	assert.InDelta(t, 10, counterState, 2)
	assert.InDelta(t, time.Millisecond*200, schedule50hz.RunTime, float64(time.Millisecond*40))
	assert.WithinRange(t, schedule50hz.LastTime, time.Now().Add(-time.Millisecond*50), time.Now())
}
