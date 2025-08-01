package ring

type Ring[T any] struct {
	buffer []T
	head   uint64
	tail   uint64
}

func NewRing[T any]() *Ring[T] {
	return &Ring[T]{
		buffer: make([]T, 1),
		head:   0,
		tail:   0,
	}
}

func (ring *Ring[T]) Head() uint64 {
	return ring.head
}

func (ring *Ring[T]) Tail() uint64 {
	return ring.tail
}

func (ring *Ring[T]) PeekHead() (element T, success bool) {
	if ring.head == ring.tail {
		return element, false
	}

	return ring.buffer[(ring.head-1)&(ring.mask())], true
}

func (ring *Ring[T]) PeekTail() (element T, success bool) {
	if ring.head == ring.tail {
		return element, false
	}

	return ring.buffer[(ring.tail)&(ring.mask())], true
}

func (ring *Ring[T]) Peek(index uint64) (element T, success bool) {
	if ring.tail > index || index >= ring.head {
		return element, false
	}

	return ring.buffer[index&(ring.mask())], true
}

func (ring *Ring[T]) DropUntil(index uint64) bool {
	if index > ring.head {
		return false
	}

	ring.tail = max(ring.tail, index)
	return true
}

func (ring *Ring[T]) Dequeue() (element T, success bool) {
	if ring.tail == ring.head {
		return element, false
	}

	element = ring.buffer[ring.tail&(ring.mask())]
	ring.tail++

	return element, true
}

func (ring *Ring[T]) EnqueueBatch(elements []T) {
	if len(elements) == 0 {
		return
	}

	for ring.Size()+uint64(len(elements)) > uint64(cap(ring.buffer)) {
		ring.grow()
	}

	startIndex := ring.head & (ring.mask())
	endIndex := (ring.head+uint64(len(elements))-1)&(ring.mask()) + 1

	if startIndex > endIndex || (startIndex == endIndex && startIndex != 0) {
		// destination looping around
		sourceLoopIndex := uint64(len(elements)) - endIndex
		copy(ring.buffer[startIndex:], elements[:sourceLoopIndex])
		copy(ring.buffer[:endIndex], elements[sourceLoopIndex:])
	} else {
		copy(ring.buffer[startIndex:endIndex], elements)
	}

	ring.head += uint64(len(elements))
}

func (ring *Ring[T]) Enqueue(element T) {
	if ring.Size() == uint64(cap(ring.buffer)) {
		ring.grow()
	}

	ring.buffer[ring.head&(ring.mask())] = element
	ring.head++
}

func (ring *Ring[T]) Size() uint64 {
	return ring.head - ring.tail
}

func (ring *Ring[T]) grow() {
	newCapacity := cap(ring.buffer) * 2
	newBuffer := make([]T, newCapacity)

	if ring.head > ring.tail {
		tailIndex := ring.tail & (ring.mask())
		headIndex := (ring.head-1)&(ring.mask()) + 1
		newTailIndex := ring.tail & uint64(newCapacity-1)
		newHeadIndex := (ring.head-1)&uint64(newCapacity-1) + 1

		if tailIndex > headIndex || (tailIndex == headIndex && tailIndex != 0) {
			if newTailIndex > newHeadIndex {
				// source and destination looping around
				copy(newBuffer[newTailIndex:newCapacity], ring.buffer[tailIndex:])
				copy(newBuffer[:newHeadIndex], ring.buffer[:headIndex])
			} else {
				// only source looping around
				copy(newBuffer[newTailIndex:cap(ring.buffer)], ring.buffer[tailIndex:])
				copy(newBuffer[cap(ring.buffer):newHeadIndex], ring.buffer[:headIndex])
			}
		} else {
			// no looping around
			copy(newBuffer[newTailIndex:newHeadIndex], ring.buffer[tailIndex:headIndex])
		}
	}

	ring.buffer = newBuffer
}

func (ring *Ring[T]) mask() uint64 {
	return uint64(cap(ring.buffer)) - 1
}
