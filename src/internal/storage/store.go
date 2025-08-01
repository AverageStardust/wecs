package storage

import (
	"sync"
)

type archetypeId uint32
type EntityId uint64
type ResourceId uint32

type Store struct {
	_            struct{} `cbor:",toarray"`
	Archetypes   []Signature
	ArchetypeMap map[uint64]archetypeId
	Parts        map[Part]struct{}
	Entries      map[EntityId]entry
	Mutex        sync.Locker
	NextEntity   EntityId
	NextTag      PartId
	Pages        map[archetypeId]*Page
	Resources    map[ResourceId]any
}

type entry struct {
	_           struct{} `cbor:",toarray"`
	ArchetypeId archetypeId
	Index       int
}

func NewStore() *Store {
	return &Store{
		Archetypes:   nil,
		ArchetypeMap: map[uint64]archetypeId{},
		Parts:        map[Part]struct{}{},
		Entries:      map[EntityId]entry{},
		Pages:        map[archetypeId]*Page{},
		Mutex:        &sync.Mutex{},
		NextEntity:   0,
		Resources:    map[ResourceId]any{},
	}
}

func (store *Store) GetComponent(entity EntityId, componentId PartId) []byte {
	entry, exists := store.Entries[entity]
	if !exists {
		return nil
	}

	page := store.Pages[entry.ArchetypeId]
	buffer, exists := page.PartBuffers[componentId]
	if !exists {
		return nil
	}

	typ := partBufferTypes[componentId]
	typeSize := int(typ.Size())
	componentOffset := entry.Index * typeSize

	return buffer[componentOffset : componentOffset+typeSize]
}

func (store *Store) NewArchetype(parts []Part) archetypeId {
	archetype := NewSignature(parts)

	hash := archetype.hash()

	id, exists := store.ArchetypeMap[hash]
	if exists {
		return id
	}

	id = archetypeId(len(store.Archetypes))

	store.Archetypes = append(store.Archetypes, archetype)
	store.ArchetypeMap[hash] = id

	return id
}

func (store *Store) move(entity EntityId, archetype archetypeId) {
	entry := store.Entries[entity]

	src := store.Pages[entry.ArchetypeId]
	dst := store.ensurePage(archetype)

	srcIndex := entry.Index
	dstIndex := dst.grow(1, entity) // grow destination page for copy

	for componentId, srcBuffer := range src.PartBuffers {
		dstBuffer, exists := dst.PartBuffers[componentId]

		// skip copy if destination does not have this component
		if !exists {
			continue
		}

		typ := partBufferTypes[componentId]
		typeSize := int(typ.Size())

		srcOffset := srcIndex * typeSize
		dstOffset := dstIndex * typeSize

		// copy component from old page to new page
		copy(dstBuffer[dstOffset:dstOffset+typeSize], srcBuffer[srcOffset:srcOffset+typeSize])
	}

	// delete entity from entries and source page
	store.Delete(entity)

	// update entry and save
	entry.ArchetypeId = archetype
	entry.Index = dstIndex
	store.Entries[entity] = entry
}

func (store *Store) Delete(entity EntityId) {
	entry, exists := store.Entries[entity]
	if !exists {
		return
	}

	page := store.Pages[entry.ArchetypeId]

	// move the entry of the last entity in the page to the deletion index
	lastEntity := page.Access[len(page.Access)-1]
	lastEntry := store.Entries[lastEntity]
	lastEntry.Index = entry.Index
	store.Entries[lastEntity] = lastEntry

	// delete the entry
	delete(store.Entries, entity)

	// delete in the page
	page.delete(entry.Index)
}

func (store *Store) Grow(archetypeId archetypeId, n int) (firstEntity EntityId) {
	firstNewIndex := store.ensurePage(archetypeId).grow(n, store.NextEntity)

	for i := range n {
		newEntity := store.NextEntity + EntityId(i)
		store.Entries[newEntity] = entry{ArchetypeId: archetypeId, Index: firstNewIndex + i}
	}

	firstEntity = store.NextEntity
	store.NextEntity += EntityId(n)

	return
}

func (store *Store) ensurePage(archetypeId archetypeId) (newPage *Page) {
	existingPage, exists := store.Pages[archetypeId]
	if exists {
		return existingPage
	}

	partBuffers := map[PartId][]byte{}
	archetype := store.Archetypes[archetypeId]
	for _, Part := range archetype {
		// record all components in use
		store.Parts[Part] = struct{}{}

		_, exists := partBufferTypes[Part.PartId()]
		if !exists {
			continue
		}

		partBuffers[Part.PartId()] = []byte{}
	}

	newPage = &Page{
		PartBuffers: partBuffers,
	}

	store.Pages[archetypeId] = newPage
	return
}
