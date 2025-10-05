package main

import (
	"reflect"
	"time"
)

// A manager to run multiple schedules on a thread.
type scheduler struct {
	_         struct{} `cbor:",toarray"`
	Schedules []*schedule
	exit      chan struct{}
}

// Create a manager to run multiple schedules on a thread.
func newScheduler() *scheduler {
	return &scheduler{
		Schedules: nil,
		exit:      nil,
	}
}

// Stop all schedules from running.
func (scheduler *scheduler) stop() bool {
	if scheduler.exit != nil {
		return false
	}

	scheduler.exit <- struct{}{}
	scheduler.exit = nil

	return true
}

// Start all schedules running again.
func (scheduler *scheduler) run(world *World) bool {
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
			schedules[chosen].run(world, time)
		} else {
			// exit
			return true
		}
	}
}

// Add a new schedule.
func (scheduler *scheduler) newSchedule(maxFrequency float64, minFrequency float64) *schedule {
	schedule := newSchedule(maxFrequency, minFrequency)
	scheduler.Schedules = append(scheduler.Schedules, schedule)

	return schedule
}

// Add a new schedule to be manually triggered.
func (scheduler *scheduler) newManualSchedule() *schedule {
	schedule := newManualSchedule()
	scheduler.Schedules = append(scheduler.Schedules, schedule)

	return schedule
}
