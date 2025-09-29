package main

import (
	"math"
	"time"

	"github.com/averagestardust/wecs/internal/storage"
)

// A schedule to run a set of systems at some interval.
type Schedule interface {
	appendSystem(system System)
	run(store *storage.Store, time time.Time)
	resetTicker()
}

type schedule struct {
	_        struct{} `cbor:",toarray"`
	ticker   *time.Ticker
	LastTime time.Time
	RunTime  time.Duration
	MinDelta time.Duration
	MaxDelta time.Duration
	Systems  []System
}

// Create a schedule that will only be run manually.
func newManualSchedule() *schedule {
	return &schedule{
		ticker:   nil,
		LastTime: time.Now(),
		RunTime:  time.Duration(0),
		MinDelta: time.Duration(0),
		MaxDelta: time.Duration(math.MaxInt64),
		Systems:  []System{},
	}
}

// Create a schedule that will attempt to run at maxFrequency, and will not allow time to pass any slower than minFrequency.
// If you want a constant time-step these arguments should be the same value.
func newSchedule(maxFrequency float64, minFrequency float64) *schedule {
	if minFrequency > maxFrequency {
		minFrequency = maxFrequency
	}

	minDelta := time.Second / time.Duration(maxFrequency)
	maxDelta := time.Second / time.Duration(minFrequency)

	schedule := &schedule{
		ticker:   time.NewTicker(minDelta),
		LastTime: time.Now(),
		RunTime:  time.Duration(0),
		MinDelta: minDelta,
		MaxDelta: maxDelta,
		Systems:  []System{},
	}

	return schedule
}

// Add a system to the schedule.
func (schedule *schedule) appendSystem(system System) {
	schedule.Systems = append(schedule.Systems, system)
}

// Run the schedule.
func (schedule *schedule) run(store *storage.Store, time time.Time) {
	delta := time.Sub(schedule.LastTime)
	schedule.LastTime = time

	delta = max(delta, schedule.MinDelta)
	delta = min(delta, schedule.MaxDelta)

	for _, system := range schedule.Systems {
		system.run(store, delta, schedule.RunTime)
	}

	schedule.RunTime += delta
}

// Reset the timer for the schedule.
func (schedule *schedule) resetTicker() {
	schedule.ticker = time.NewTicker(schedule.MinDelta)
	schedule.LastTime = time.Now()
}
