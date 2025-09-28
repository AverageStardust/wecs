package storage

import (
	"iter"
	"slices"
)

type Page struct {
	_           struct{} `cbor:",toarray"`
	PartBuffers map[PartId][]byte
	Entities    []EntityId
	DirtySize   int
	Size        int
}

func (page *Page) GetComponentIter(componentId PartId) iter.Seq[[]byte] {
	typ, success := partBufferTypes[componentId]

	if !success {
		// empty iterator if this page doesn't have that component
		return func(yield func([]byte) bool) {}
	}

	buffer := page.PartBuffers[componentId]
	typeSize := int(typ.Size())

	return func(yield func([]byte) bool) {
		for i := range page.Size {
			componentOffset := i * typeSize
			componentBytes := buffer[componentOffset : componentOffset+typeSize]
			if !yield(componentBytes) {
				return
			}
		}
	}
}

// deletes index in the page and moves last element to fill it's place
func (page *Page) delete(index int) {
	lastIndex := len(page.Entities) - 1

	// move the last entity to the deletion location
	page.Entities[index] = page.Entities[lastIndex]

	// delete the last entity
	page.Entities = page.Entities[:lastIndex]

	// last we must move all the components of the last entity
	// now that the last element has been written over the removed element, we can shorten
	for componentId, buffer := range page.PartBuffers {
		typ := partBufferTypes[componentId]
		typeSize := int(typ.Size())

		indexOffset := index * typeSize
		lastIndexOffset := lastIndex * typeSize

		// move the component of the last entity to the deletion index
		copy(buffer[indexOffset:indexOffset+typeSize], buffer[lastIndexOffset:lastIndexOffset+typeSize])

		// release the last bytes of the page
		page.PartBuffers[componentId] = buffer[0 : len(buffer)-typeSize]
	}

	page.Size--
}

func (page *Page) grow(n int, firstEntity EntityId) (firstIndex int) {
	// grow component buffers
	for componentId, buffer := range page.PartBuffers {
		typ := partBufferTypes[componentId]
		typeSize := int(typ.Size())
		growSize := typeSize * n
		newLen := len(buffer) + growSize

		buffer = slices.Grow(buffer, growSize) // grow capacity

		if page.DirtySize > page.Size {
			// clear if underlying array is dirty
			clear(buffer[len(buffer):min(newLen, page.DirtySize*typeSize)])
		}

		buffer = buffer[0:newLen] // grow length

		page.PartBuffers[componentId] = buffer
	}

	firstIndex = len(page.Entities)

	// grow entity list
	for i := range n {
		page.Entities = append(page.Entities, firstEntity+EntityId(i))
	}

	page.Size += n
	page.DirtySize = max(page.DirtySize, page.Size)
	return
}
