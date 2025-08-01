package storage

import (
	"iter"
	"slices"
)

type Page struct {
	_           struct{} `cbor:",toarray"`
	PartBuffers map[PartId][]byte
	Access      []EntityId
	DirtySize   int
	Size        int
}

func (page *Page) GetComponentIter(componentId PartId) iter.Seq[[]byte] {
	buffer := page.PartBuffers[componentId]
	typ := partBufferTypes[componentId]
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
	lastIndex := len(page.Access) - 1

	// move the last entity to the deletion location
	page.Access[index] = page.Access[lastIndex]

	// delete the last entity
	page.Access = page.Access[:lastIndex]

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

	firstIndex = len(page.Access)

	// grow entity list
	for i := range n {
		page.Access = append(page.Access, firstEntity+EntityId(i))
	}

	page.Size += n
	page.DirtySize = max(page.DirtySize, page.Size)
	return
}
