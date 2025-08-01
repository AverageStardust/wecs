package scheduler

import (
	"time"

	"github.com/averagestardust/wecs/internal/storage"
)

type SystemId uint32

type Schedule struct {
	_            struct{} `cbor:",toarray"`
	ticker       *time.Ticker
	LastTime     time.Time
	RunTime      time.Duration
	MinDelta     time.Duration
	MaxDelta     time.Duration
	SystemStates map[SystemId]systemRunner
}

type systemRunner interface {
	Run(systemId SystemId, store *storage.Store, delta, runtime time.Duration)
}

func NewSchedule(maxFrequency float64, minFrequency float64) *Schedule {
	if minFrequency > maxFrequency {
		minFrequency = maxFrequency
	}

	minDelta := time.Second / time.Duration(maxFrequency)
	maxDelta := time.Second / time.Duration(minFrequency)

	schedule := &Schedule{
		ticker:       time.NewTicker(minDelta),
		LastTime:     time.Now(),
		RunTime:      time.Duration(0),
		MinDelta:     minDelta,
		MaxDelta:     maxDelta,
		SystemStates: map[SystemId]systemRunner{},
	}

	return schedule
}

func (schedule *Schedule) run(store *storage.Store, time time.Time) {
	delta := time.Sub(schedule.LastTime)
	schedule.LastTime = time

	delta = max(delta, schedule.MinDelta)
	delta = min(delta, schedule.MaxDelta)
	schedule.RunTime += delta

	for systemId, systemState := range schedule.SystemStates {
		systemState.Run(systemId, store, delta, schedule.RunTime)
	}
}

func (schedule *Schedule) ResetTicker() {
	schedule.ticker = time.NewTicker(schedule.MinDelta)
	schedule.LastTime = time.Now()
}
