package ring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRingGrow(t *testing.T) {
	ring0 := &Ring[int]{
		buffer: []int{4, 1, 2, 3},
		head:   5,
		tail:   1,
	}

	ring0.grow()
	assert.EqualValues(t,
		Ring[int]{
			buffer: []int{0, 1, 2, 3, 4, 0, 0, 0},
			head:   5,
			tail:   1,
		}, *ring0)

	ring1 := &Ring[int]{
		buffer: []int{7, 8},
		head:   5,
		tail:   3,
	}

	ring1.grow()
	assert.EqualValues(t,
		Ring[int]{
			buffer: []int{7, 0, 0, 8},
			head:   5,
			tail:   3,
		}, *ring1)

	ring1.grow()
	assert.EqualValues(t,
		Ring[int]{
			buffer: []int{0, 0, 0, 8, 7, 0, 0, 0},
			head:   5,
			tail:   3,
		}, *ring1)

	ring2 := &Ring[int]{
		buffer: []int{6745, 2},
		head:   4,
		tail:   2,
	}

	ring2.grow()
	assert.EqualValues(t,
		Ring[int]{
			buffer: []int{0, 0, 6745, 2},
			head:   4,
			tail:   2,
		}, *ring2)
}
