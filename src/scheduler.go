package main

import (
	"reflect"
	"time"

	"github.com/averagestardust/wecs/internal/storage"
)

type scheduler struct {
	_         struct{} `cbor:",toarray"`
	Schedules []*Schedule
	exit      chan struct{}
}

func newScheduler() *scheduler {
	return &scheduler{
		Schedules: nil,
		exit:      nil,
	}
}

func (scheduler *scheduler) stop() bool {
	if scheduler.exit != nil {
		return false
	}

	scheduler.exit <- struct{}{}
	scheduler.exit = nil

	return true
}

func (scheduler *scheduler) run(store *storage.Store) bool {
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

func (scheduler *scheduler) newSchedule(maxFrequency float64, minFrequency float64) *Schedule {
	schedule := newSchedule(maxFrequency, minFrequency)
	scheduler.Schedules = append(scheduler.Schedules, schedule)

	return schedule
}

func (scheduler *scheduler) newManuelSchedule() *Schedule {
	schedule := newManuelSchedule()
	scheduler.Schedules = append(scheduler.Schedules, schedule)

	return schedule
}
