package ring_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/averagestardust/wecs/internal/ring"
	"github.com/stretchr/testify/assert"
)

func FuzzRing(f *testing.F) {
	f.Fuzz(fuzzRingTarget)
}

func fuzzRingTarget(t *testing.T, seedA, seedB uint64) {
	rng := rand.New(rand.NewPCG(seedA, seedB))
	size := rng.IntN(4096)
	data := make([]int, size)

	for i := range size {
		data[i] = rng.Int()
	}

	added := 0
	removed := 0
	ring := ring.NewRing[int]()

	for removed < size {
		if added < size && rng.IntN(2) == 1 {
			if rng.IntN(3) < 2 {
				fmt.Println("enqueue ", added)
				ring.Enqueue(data[added])
				added++
			} else {
				amount := rng.IntN(size - added + 1)
				fmt.Println("enqueue batch ", added, added+amount)
				ring.EnqueueBatch(data[added : added+amount])
				added += amount
			}
		} else {
			if rng.IntN(3) < 2 {
				fmt.Println("de ", removed)
				got, success := ring.Dequeue()
				assert.Equal(t, added > removed, success)

				if added > removed {
					assert.Equal(t, data[removed], got)
					removed++
				}
			} else {
				index := rng.IntN(added + 1)
				fmt.Println("drop until ", index)
				success := ring.DropUntil(uint64(index))
				assert.Equal(t, index <= added, success)

				removed = max(removed, index)
			}
		}

		for i := removed; i < added; i++ {
			elm, success := ring.Peek(uint64(i))
			if data[i] != elm || !success {
				print("@@@", i, "\n")
				break
			}
		}
	}
}

func TestRingHead(t *testing.T) {
	ring := ring.NewRing[int]()

	assert.Equal(t, uint64(0), ring.Head())

	ring.Enqueue(3)
	ring.Enqueue(6)

	assert.Equal(t, uint64(2), ring.Head())

	ring.EnqueueBatch([]int{5, 8, 12, -3})

	assert.Equal(t, uint64(6), ring.Head())

	ring.Dequeue()

	assert.Equal(t, uint64(6), ring.Head())

	ring.DropUntil(6)

	assert.Equal(t, uint64(6), ring.Head())
}

func TestRingTail(t *testing.T) {
	ring := ring.NewRing[int]()

	assert.Equal(t, uint64(0), ring.Tail())

	ring.Enqueue(3)
	ring.Enqueue(6)

	assert.Equal(t, uint64(0), ring.Tail())

	ring.EnqueueBatch([]int{5, 8, 12, -3})

	assert.Equal(t, uint64(0), ring.Tail())

	ring.Dequeue()

	assert.Equal(t, uint64(1), ring.Tail())

	ring.DropUntil(6)

	assert.Equal(t, uint64(6), ring.Tail())
}

func TestRingPeekHead(t *testing.T) {
	ring := ring.NewRing[int]()

	_, success := ring.PeekHead()
	assert.False(t, success)

	ring.Enqueue(6)
	ring.Enqueue(23)

	got, success := ring.PeekHead()
	assert.True(t, success)
	assert.Equal(t, 23, got)

	ring.EnqueueBatch([]int{5, -3, 93})

	got, success = ring.PeekHead()
	assert.True(t, success)
	assert.Equal(t, 93, got)

	ring.Dequeue()
	ring.Dequeue()
	ring.Dequeue()

	got, success = ring.PeekHead()
	assert.True(t, success)
	assert.Equal(t, 93, got)

	ring.DropUntil(5)

	_, success = ring.PeekHead()
	assert.False(t, success)
}

func TestRingPeekTail(t *testing.T) {
	ring := ring.NewRing[int]()

	_, success := ring.PeekTail()
	assert.False(t, success)

	ring.Enqueue(6)
	ring.Enqueue(23)

	got, success := ring.PeekTail()
	assert.True(t, success)
	assert.Equal(t, 6, got)

	ring.EnqueueBatch([]int{5, -3, 93})

	got, success = ring.PeekTail()
	assert.True(t, success)
	assert.Equal(t, 6, got)

	ring.Dequeue()
	ring.Dequeue()
	ring.Dequeue()

	got, success = ring.PeekTail()
	assert.True(t, success)
	assert.Equal(t, -3, got)

	ring.DropUntil(5)

	_, success = ring.PeekTail()
	assert.False(t, success)
}

func TestRingPeek(t *testing.T) {
	ring := ring.NewRing[int]()

	assert.True(t, ring.DropUntil(0))
	assert.False(t, ring.DropUntil(1))

	_, success := ring.Peek(0)
	assert.False(t, success)

	_, success = ring.Peek(5)
	assert.False(t, success)

	ring.EnqueueBatch([]int{6, 23, -3, 93})
	ring.Dequeue()

	_, success = ring.Peek(0)
	assert.False(t, success)

	got, success := ring.Peek(1)
	assert.True(t, success)
	assert.Equal(t, 23, got)

	got, success = ring.Peek(3)
	assert.True(t, success)
	assert.Equal(t, 93, got)

	_, success = ring.Peek(4)
	assert.False(t, success)
}

func TestRingDropUntil(t *testing.T) {
	ring := ring.NewRing[int]()

	_, success := ring.Peek(0)
	assert.False(t, success)

	_, success = ring.Peek(5)
	assert.False(t, success)

	ring.EnqueueBatch([]int{6, 23, -3, 93})

	assert.True(t, ring.DropUntil(2))

	_, success = ring.Peek(1)
	assert.False(t, success)

	got, success := ring.Peek(2)
	assert.True(t, success)
	assert.Equal(t, -3, got)

	got, success = ring.Peek(3)
	assert.True(t, success)
	assert.Equal(t, 93, got)

	_, success = ring.Peek(4)
	assert.False(t, success)

	assert.True(t, ring.DropUntil(4))

	_, success = ring.Peek(3)
	assert.False(t, success)

	assert.False(t, ring.DropUntil(5))
}

func TestRingDequeue(t *testing.T) {
	ring := ring.NewRing[int]()

	_, success := ring.Dequeue()
	assert.False(t, success)

	ring.EnqueueBatch([]int{6, 23})

	got, success := ring.Dequeue()
	assert.True(t, success)
	assert.Equal(t, 6, got)

	got, success = ring.Dequeue()
	assert.True(t, success)
	assert.Equal(t, 23, got)

	_, success = ring.Dequeue()
	assert.False(t, success)
}

func TestRingEnqueueBatch(t *testing.T) {
	ring := ring.NewRing[int]()

	ring.EnqueueBatch([]int{6, 23})
	ring.EnqueueBatch([]int{745, 34})

	got, success := ring.Peek(0)
	assert.True(t, success)
	assert.Equal(t, 6, got)

	got, success = ring.Peek(3)
	assert.True(t, success)
	assert.Equal(t, 34, got)

	_, success = ring.Peek(4)
	assert.False(t, success)

	ring.EnqueueBatch([]int{86})
	ring.EnqueueBatch([]int{})
	ring.EnqueueBatch([]int{934, 49, -3})

	got, success = ring.Peek(4)
	assert.True(t, success)
	assert.Equal(t, 86, got)

	got, success = ring.Peek(5)
	assert.True(t, success)
	assert.Equal(t, 934, got)

	got, success = ring.Peek(7)
	assert.True(t, success)
	assert.Equal(t, -3, got)

	_, success = ring.Peek(8)
	assert.False(t, success)
}

func TestRingEnqueue(t *testing.T) {
	ring := ring.NewRing[int]()

	ring.Enqueue(4)
	ring.Enqueue(-63)
	ring.Enqueue(-96)
	ring.Enqueue(129)

	got, success := ring.Peek(0)
	assert.True(t, success)
	assert.Equal(t, 4, got)

	got, success = ring.Peek(2)
	assert.True(t, success)
	assert.Equal(t, -96, got)

	got, success = ring.Peek(3)
	assert.True(t, success)
	assert.Equal(t, 129, got)

	_, success = ring.Peek(4)
	assert.False(t, success)
}

func TestRingSize(t *testing.T) {
	ring := ring.NewRing[int]()

	assert.Equal(t, uint64(0), ring.Size())

	ring.Enqueue(4)
	ring.Enqueue(-63)
	ring.Enqueue(-96)

	assert.Equal(t, uint64(3), ring.Size())

	ring.Dequeue()
	ring.Dequeue()

	assert.Equal(t, uint64(1), ring.Size())
}
