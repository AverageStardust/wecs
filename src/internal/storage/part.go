package storage

import (
	"encoding/binary"
	"hash/crc64"
	"reflect"

	"github.com/averagestardust/wecs/internal/common"
)

type PartId uint32
type Part interface {
	PartId() PartId
}

var partBufferTypes = map[PartId]reflect.Type{}

func cmpPart(a Part, b Part) int {
	return int(a.PartId()) - int(b.PartId())
}

func NewPartType(partId PartId, typ reflect.Type) {
	partBufferTypes[partId] = typ
}

func (store *Store) DeletePart(entity EntityId, part Part) (success bool) {
	entry, exists := store.Entries[entity]

	if !exists {
		return false
	}

	oldArchetypeId := entry.ArchetypeId
	archetype := store.Archetypes[oldArchetypeId].Delete(part)
	archetypeId := store.NewArchetype(archetype)

	if oldArchetypeId == archetypeId {
		// failed because entity never had that part
		return false
	}

	store.move(entity, archetypeId)
	return true
}

func (store *Store) HasPart(entity EntityId, part Part) (has bool) {
	entry, exists := store.Entries[entity]
	if !exists {
		return false
	}

	return store.Archetypes[entry.ArchetypeId].ContainsSingle(part)
}

func (store *Store) AddPart(entity EntityId, part Part) (success bool) {
	entry, exists := store.Entries[entity]

	if !exists {
		return false
	}

	oldArchetypeId := entry.ArchetypeId
	oldArchetype := store.Archetypes[oldArchetypeId]
	archetypeId := store.NewArchetype(oldArchetype.Add(part))

	if oldArchetypeId == archetypeId {
		// failed because entity already had that part
		return false
	}

	store.move(entity, archetypeId)
	return true
}

func HashUsedParts(store *Store) uint64 {
	hash := crc64.New(common.Crc64ISOTable)
	uint32Bytes := make([]byte, 4)

	for part := range store.Parts {
		componentId := part.PartId()
		binary.LittleEndian.PutUint32(uint32Bytes, uint32(componentId))
		hash.Write(uint32Bytes)

		if int(componentId) >= len(partBufferTypes) {
			continue
		}

		typ, exists := partBufferTypes[componentId]
		if !exists {
			continue
		}

		hash.Write([]byte(typ.Name()))
	}

	return hash.Sum64()
}
