package main

import (
	"math"
	"time"

	"github.com/averagestardust/wecs/internal/storage"
)

type Schedule struct {
	_        struct{} `cbor:",toarray"`
	ticker   *time.Ticker
	LastTime time.Time
	RunTime  time.Duration
	MinDelta time.Duration
	MaxDelta time.Duration
	Systems  []System
}

func newManuelSchedule() *Schedule {
	return &Schedule{
		ticker:   nil,
		LastTime: time.Now(),
		RunTime:  time.Duration(0),
		MinDelta: time.Duration(0),
		MaxDelta: time.Duration(math.MaxInt64),
		Systems:  []System{},
	}
}

func newSchedule(maxFrequency float64, minFrequency float64) *Schedule {
	if minFrequency > maxFrequency {
		minFrequency = maxFrequency
	}

	minDelta := time.Second / time.Duration(maxFrequency)
	maxDelta := time.Second / time.Duration(minFrequency)

	schedule := &Schedule{
		ticker:   time.NewTicker(minDelta),
		LastTime: time.Now(),
		RunTime:  time.Duration(0),
		MinDelta: minDelta,
		MaxDelta: maxDelta,
		Systems:  []System{},
	}

	return schedule
}

func (schedule *Schedule) run(store *storage.Store, time time.Time) {
	delta := time.Sub(schedule.LastTime)
	schedule.LastTime = time

	delta = max(delta, schedule.MinDelta)
	delta = min(delta, schedule.MaxDelta)
	schedule.RunTime += delta

	for _, system := range schedule.Systems {
		system.run(store, delta, schedule.RunTime)
	}
}

func (schedule *Schedule) resetTicker() {
	schedule.ticker = time.NewTicker(schedule.MinDelta)
	schedule.LastTime = time.Now()
}
