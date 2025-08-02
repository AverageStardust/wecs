package scheduler

import (
	"math"
	"time"

	"github.com/averagestardust/wecs/internal/storage"
)

type SystemId uint32

type Schedule struct {
	_        struct{} `cbor:",toarray"`
	ticker   *time.Ticker
	LastTime time.Time
	RunTime  time.Duration
	MinDelta time.Duration
	MaxDelta time.Duration
	Systems  map[SystemId]systemRunner
}

type systemRunner interface {
	Run(systemId SystemId, store *storage.Store, delta, runtime time.Duration)
}

func NewSchedule(maxFrequency float64, minFrequency float64) *Schedule {
	if maxFrequency == 0 {
		return &Schedule{
			ticker:   nil,
			LastTime: time.Now(),
			RunTime:  time.Duration(0),
			MinDelta: 0,
			MaxDelta: math.MaxInt64,
			Systems:  map[SystemId]systemRunner{},
		}
	}

	if minFrequency > maxFrequency {
		minFrequency = maxFrequency
	}

	if minFrequency < 1e-6 {
		minFrequency = 1e-6
		maxFrequency = 1e-6
	}

	minDelta := time.Second / time.Duration(maxFrequency)
	maxDelta := time.Second / time.Duration(minFrequency)

	schedule := &Schedule{
		ticker:   time.NewTicker(minDelta),
		LastTime: time.Now(),
		RunTime:  time.Duration(0),
		MinDelta: minDelta,
		MaxDelta: maxDelta,
		Systems:  map[SystemId]systemRunner{},
	}

	return schedule
}

func (schedule *Schedule) Run(store *storage.Store, time time.Time) {
	delta := time.Sub(schedule.LastTime)
	schedule.LastTime = time

	delta = max(delta, schedule.MinDelta)
	delta = min(delta, schedule.MaxDelta)
	schedule.RunTime += delta

	for systemId, system := range schedule.Systems {
		system.Run(systemId, store, delta, schedule.RunTime)
	}
}

func (schedule *Schedule) ResetTicker() {
	schedule.ticker = time.NewTicker(schedule.MinDelta)
	schedule.LastTime = time.Now()
}
