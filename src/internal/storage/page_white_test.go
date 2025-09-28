package storage

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPageGetComponentIter(t *testing.T) {
	Page := newTestPage(
		[]EntityId{54, 9, 32},
		[]byte{0, 0, 4, 0, 25, 1},                  // []uint16{0, 4, 281}
		[]byte{3, 0, 0, 0, 4, 1, 0, 0, 2, 0, 1, 0}) // []uint32{3, 260, 65538}

	shorts := []uint16{}
	for bytes := range Page.GetComponentIter(0) {
		shorts = append(shorts, binary.LittleEndian.Uint16(bytes))
	}

	assert.ElementsMatch(t, shorts, []uint16{0, 4, 281})

	integers := []uint32{}
	for bytes := range Page.GetComponentIter(1) {
		integers = append(integers, binary.LittleEndian.Uint32(bytes))
	}

	assert.ElementsMatch(t, integers, []uint32{3, 260, 65538})
}

func TestPageDelete(t *testing.T) {
	Page := newTestPage(
		[]EntityId{54, 9, 32},
		[]byte{0, 0, 4, 0, 23, 1},                  // []uint16{0, 4, 279}
		[]byte{3, 0, 0, 0, 4, 1, 0, 0, 0, 0, 0, 0}) // []uint32{3, 260, 0}

	Page.delete(1)

	assert.ElementsMatch(t, []entityData{{54, 0, 3}, {32, 279, 0}}, readPage(Page))
}

func TestPageGrow(t *testing.T) {
	Page := newTestPage(
		[]EntityId{12},
		[]byte{3, 1},       // []uint16{259}
		[]byte{6, 0, 0, 0}) // []uint32{6}

	Page.grow(2, EntityId(25))

	assert.ElementsMatch(t, []entityData{{12, 259, 6}, {25, 0, 0}, {26, 0, 0}}, readPage(Page))
}

func newTestPage(entityIds []EntityId, shortBuffer, integerBuffer []byte) *Page {
	partBufferTypes = map[PartId]reflect.Type{
		PartId(shortComponent):   reflect.TypeFor[uint16](),
		PartId(integerComponent): reflect.TypeFor[uint32](),
	}

	return &Page{
		PartBuffers: map[PartId][]byte{
			PartId(shortComponent):   shortBuffer,
			PartId(integerComponent): integerBuffer,
		},
		Entities:  entityIds,
		Size:      len(entityIds),
		DirtySize: len(entityIds),
	}
}

func readPage(Page *Page) []entityData {
	entities := []entityData{}

	for i, entity := range Page.Entities {
		short := binary.LittleEndian.Uint16(Page.PartBuffers[0][i*2 : i*2+2])
		integer := binary.LittleEndian.Uint32(Page.PartBuffers[1][i*4 : i*4+4])
		entities = append(entities, entityData{entity, short, integer})
	}

	return entities
}
