package scheduler_test

import (
	"testing"
	"time"

	"github.com/averagestardust/wecs/internal/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestNewSchedule(t *testing.T) {
	scheduleA := scheduler.NewSchedule(50, 10)
	assert.Equal(t, time.Millisecond*20, scheduleA.MinDelta)
	assert.Equal(t, time.Millisecond*100, scheduleA.MaxDelta)

	scheduleB := scheduler.NewSchedule(20, 50)
	assert.Equal(t, time.Millisecond*50, scheduleB.MinDelta)
	assert.Equal(t, time.Millisecond*50, scheduleB.MaxDelta)
}
