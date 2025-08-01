package scheduler

import (
	"reflect"
	"time"

	"github.com/averagestardust/wecs/internal/storage"
)

type Scheduler struct {
	_         struct{} `cbor:",toarray"`
	Schedules []*Schedule
	exit      chan struct{}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		Schedules: nil,
		exit:      nil,
	}
}

func (scheduler *Scheduler) StopSystems() bool {
	if scheduler.exit != nil {
		return false
	}

	scheduler.exit <- struct{}{}
	scheduler.exit = nil

	return true
}

func (scheduler *Scheduler) RunSystems(store *storage.Store) bool {
	if scheduler.exit != nil {
		return false
	}
	scheduler.exit = make(chan struct{})

	schedules := scheduler.Schedules
	cases := []reflect.SelectCase{}

	for _, schedule := range schedules {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(schedule.ticker.C),
		})
	}

	cases = append(cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(scheduler.exit),
	})

	for {
		chosen, received, _ := reflect.Select(cases)
		if chosen < len(schedules) {
			time := received.Interface().(time.Time)
			schedules[chosen].run(store, time)
		} else {
			// exit
			return true
		}
	}
}

func (scheduler *Scheduler) NewSchedule(frequency float64) *Schedule {
	schedule := NewSchedule(frequency, frequency)
	scheduler.Schedules = append(scheduler.Schedules, schedule)

	return schedule
}

func (scheduler *Scheduler) NewVariableSchedule(maxFrequency float64, minFrequency float64) *Schedule {
	schedule := NewSchedule(maxFrequency, minFrequency)
	scheduler.Schedules = append(scheduler.Schedules, schedule)

	return schedule
}
