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
	Systems  []systemRunner
}

type systemRunner interface {
	Run(store *storage.Store, delta, runtime time.Duration)
	Id() SystemId
}

func NewSchedule(maxFrequency float64, minFrequency float64) *Schedule {
	if maxFrequency == 0 {
		return &Schedule{
			ticker:   nil,
			LastTime: time.Now(),
			RunTime:  time.Duration(0),
			MinDelta: 0,
			MaxDelta: math.MaxInt64,
			Systems:  []systemRunner{},
		}
	}

func NewSchedule(maxFrequency float64, minFrequency float64) *Schedule {
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
		Systems:  []systemRunner{},
	}

	return schedule
}

func (schedule *Schedule) Run(store *storage.Store, time time.Time) {
	delta := time.Sub(schedule.LastTime)
	schedule.LastTime = time

	delta = max(delta, schedule.MinDelta)
	delta = min(delta, schedule.MaxDelta)
	schedule.RunTime += delta

	for _, system := range schedule.Systems {
		system.Run(store, delta, schedule.RunTime)
	}
}

func (schedule *Schedule) Delete(system systemRunner) {
	keptRunners := []systemRunner{}
	id := system.Id()

	for _, runner := range schedule.Systems {
		if id == runner.Id() {
			keptRunners = append(keptRunners, runner)
		}
	}

	schedule.Systems = keptRunners
}

func (schedule *Schedule) Has(system systemRunner) bool {
	id := system.Id()

	for _, runner := range schedule.Systems {
		if id == runner.Id() {
			return true
		}
	}

	return false
}

func (schedule *Schedule) Add(system systemRunner) {
	schedule.Systems = append(schedule.Systems, system)
}

func (schedule *Schedule) ResetTicker() {
	schedule.ticker = time.NewTicker(schedule.MinDelta)
	schedule.LastTime = time.Now()
}
