package storage

import (
	"encoding/binary"
	"hash/crc64"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type entityData struct {
	id      EntityId
	short   uint16
	integer uint32
}

var shortComponent = partMock(0)
var integerComponent = partMock(1)

func TestStorageGetComponent(t *testing.T) {
	storage := newTestStore(
		map[EntityId]entry{
			0: {ArchetypeId: 2, Index: 0},
			2: {ArchetypeId: 2, Index: 1},
			3: {ArchetypeId: 2, Index: 2},
		},
		map[archetypeId]*Page{
			2: newTestPage(
				[]EntityId{0, 2, 3},
				[]byte{9, 0, 34, 1, 78, 0},
				[]byte{0, 1, 0, 0, 3, 53, 230, 1, 7, 0, 0, 0}),
		}, 4)

	got := storage.GetComponent(EntityId(0), 0)
	if assert.NotNil(t, got) {
		assert.Equal(t, uint16(9), binary.LittleEndian.Uint16(got))
		assert.Equal(t, 2, len(got))
	}

	got = storage.GetComponent(EntityId(2), 0)
	if assert.NotNil(t, got) {
		assert.Equal(t, uint16(290), binary.LittleEndian.Uint16(got))
		assert.Equal(t, 2, len(got))
	}

	got = storage.GetComponent(EntityId(3), 0)
	if assert.NotNil(t, got) {
		assert.Equal(t, uint16(78), binary.LittleEndian.Uint16(got))
		assert.Equal(t, 2, len(got))
	}

	got = storage.GetComponent(EntityId(0), 1)
	if assert.NotNil(t, got) {
		assert.Equal(t, uint32(256), binary.LittleEndian.Uint32(got))
		assert.Equal(t, 4, len(got))
	}

	got = storage.GetComponent(EntityId(2), 1)
	if assert.NotNil(t, got) {
		assert.Equal(t, uint32(31864067), binary.LittleEndian.Uint32(got))
		assert.Equal(t, 4, len(got))
	}

	got = storage.GetComponent(EntityId(3), 1)
	if assert.NotNil(t, got) {
		assert.Equal(t, uint32(7), binary.LittleEndian.Uint32(got))
		assert.Equal(t, 4, len(got))
	}
}

func TestStorageNewArchetype(t *testing.T) {
	tag0 := partMock(^uint32(0))
	tag2 := partMock(^uint32(2))
	component1 := partMock(1)
	component5 := partMock(5)

	storage := NewStore()

	assert.Equal(t, archetypeId(0), storage.NewArchetype([]Part{tag0}))
	assert.Equal(t, archetypeId(1), storage.NewArchetype([]Part{tag2, component1}))
	assert.Equal(t, archetypeId(2), storage.NewArchetype([]Part{component1, component5}))

	// repeat
	assert.Equal(t, archetypeId(1), storage.NewArchetype([]Part{component1, tag2}))

	assert.Equal(t, archetypeId(3), storage.NewArchetype([]Part{}))
	assert.Equal(t, archetypeId(4), storage.NewArchetype([]Part{tag0, tag2, component1, component5}))

	// repeat
	assert.Equal(t, archetypeId(3), storage.NewArchetype([]Part{}))
}

func TestStorageMove(t *testing.T) {
	storage := newTestStore(
		map[EntityId]entry{
			0: {ArchetypeId: 1, Index: 0},
			2: {ArchetypeId: 2, Index: 0},
			3: {ArchetypeId: 2, Index: 1},
		},
		map[archetypeId]*Page{
			1: {
				PartBuffers: map[PartId][]byte{},
				Entities:    []EntityId{0},
				Size:        1,
				DirtySize:   1,
			},
			2: newTestPage(
				[]EntityId{2, 3},
				[]byte{34, 1, 78, 0},
				[]byte{3, 53, 230, 1, 7, 0, 0, 0}),
		}, 4)

	// drop integer component from entity 2
	storage.move(2, 0)

	assert.EqualValues(t,
		Page{
			PartBuffers: map[PartId][]byte{
				PartId(shortComponent): {34, 1},
			},
			Entities:  []EntityId{2},
			Size:      1,
			DirtySize: 1,
		},
		*storage.Pages[0])

	assert.EqualValues(t,
		Page{
			PartBuffers: map[PartId][]byte{
				PartId(shortComponent):   {78, 0},
				PartId(integerComponent): {7, 0, 0, 0},
			},
			Entities:  []EntityId{3},
			Size:      1,
			DirtySize: 2,
		},
		*storage.Pages[2])

	assert.EqualValues(t,
		map[EntityId]entry{
			0: {ArchetypeId: 1, Index: 0},
			2: {ArchetypeId: 0, Index: 0},
			3: {ArchetypeId: 2, Index: 0},
		},
		storage.Entries)

	// add short and integer components to entity 0
	storage.move(0, 2)

	assert.EqualValues(t,
		Page{
			PartBuffers: map[PartId][]byte{},
			Entities:    []EntityId{},
			Size:        0,
			DirtySize:   1,
		},
		*storage.Pages[1])

	assert.ElementsMatch(t, []entityData{{0, 0, 0}, {3, 78, 7}}, readPage(storage.Pages[2]))

	assert.Equal(t, storage.Entries[0].ArchetypeId, archetypeId(2))
	assert.Equal(t, storage.Entries[2].ArchetypeId, archetypeId(0))
	assert.Equal(t, storage.Entries[3].ArchetypeId, archetypeId(2))

	assert.NotEqual(t, storage.Entries[0].Index, storage.Entries[3].Index)
}

func TestStorageGrow(t *testing.T) {
	storage := newTestStore(
		map[EntityId]entry{
			4: {ArchetypeId: 2, Index: 0},
		},
		map[archetypeId]*Page{
			2: newTestPage(
				[]EntityId{4},
				[]byte{0, 0},
				[]byte{0, 0, 0, 0}),
		}, 7)

	storage.Grow(0, 2)
	assert.EqualValues(t,
		map[EntityId]entry{
			4: {ArchetypeId: 2, Index: 0},

			7: {ArchetypeId: 0, Index: 0},
			8: {ArchetypeId: 0, Index: 1}},
		storage.Entries)

	storage.Grow(2, 3)
	assert.EqualValues(t,
		map[EntityId]entry{
			4: {ArchetypeId: 2, Index: 0},

			7: {ArchetypeId: 0, Index: 0},
			8: {ArchetypeId: 0, Index: 1},

			9:  {ArchetypeId: 2, Index: 1},
			10: {ArchetypeId: 2, Index: 2},
			11: {ArchetypeId: 2, Index: 3}},
		storage.Entries)
}

func TestStorageEnsurePage(t *testing.T) {
	storage := newTestStore(
		map[EntityId]entry{},
		map[archetypeId]*Page{
			2: {},
		}, 0)

	originalPage := storage.Pages[2]

	got2 := storage.ensurePage(2)
	assert.NotNil(t, storage.Pages[2])
	assert.Equal(t, originalPage, got2)

	got0 := storage.ensurePage(0)
	assert.NotNil(t, storage.Pages[0])
	assert.Equal(t, Page{
		PartBuffers: map[PartId][]byte{
			PartId(shortComponent): {}, // short component in archetype 0 from newTestStore()
		},
	}, *got0)

	got1 := storage.ensurePage(1)
	assert.NotNil(t, storage.Pages[1])
	assert.Equal(t, Page{
		PartBuffers: map[PartId][]byte{}, // archetype 1 from newTestStore() is all tags
	}, *got1)
}

func newTestStore(entries map[EntityId]entry, Pages map[archetypeId]*Page, nextEntity EntityId) *Store {
	tag1 := partMock(^uint32(1))
	tag3 := partMock(^uint32(3))

	isoTable := crc64.MakeTable(crc64.ISO)
	crc64.Checksum(
		[]byte{1, 0, 0, 0, 5, 0, 0, 0, 254, 255, 255, 255}, isoTable)

	return &Store{
		Archetypes: []Signature{
			{shortComponent, tag1},
			{tag3},
			{shortComponent, integerComponent},
		},
		ArchetypeMap: map[uint64]archetypeId{
			crc64.Checksum([]byte{0, 0, 0, 0, 254, 255, 255, 255}, isoTable): 0,
			crc64.Checksum([]byte{252, 255, 255, 255}, isoTable):             1,
			crc64.Checksum([]byte{0, 0, 0, 0, 1, 0, 0, 0}, isoTable):         2,
		},
		Parts: map[Part]struct{}{
			shortComponent:   {},
			integerComponent: {},
		},
		Entries:    entries,
		Pages:      Pages,
		Mutex:      &sync.Mutex{},
		NextEntity: nextEntity,
		Resources:  map[string]any{},
	}
}
