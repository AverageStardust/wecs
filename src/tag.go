package main

import (
	"github.com/averagestardust/wecs/internal/storage"
)

// An integer uniquely identifying a tag type.
// Tags should identify boolean properties entity.
// Should be created in a static order during world initialization.
type Tag storage.PartId

// The next tag id to be assigned when a component is created.
// Starts from the maximum uint32 value and decrements.
var nextTag uint32 = ^uint32(0)

// Creates a new type of tag to identify boolean properties entities.
// Should be used in a static order during world initialization.
func NewTag() Tag {
	tag := Tag(nextTag)
	nextTag--

	return tag
}

// Remove a tag from an entity.
func (tag Tag) Delete(access *Access, entity Entity) (success bool) {
	return access.store.DeletePart(storage.EntityId(entity), tag)
}

// Check if an entity has a tag.
func (tag Tag) Has(access *Access, entity Entity) (has bool) {
	return access.store.HasPart(storage.EntityId(entity), tag)
}

// Add a tag to an entity.
func (tag Tag) Add(access *Access, entity Entity) (success bool) {
	return access.store.AddPart(storage.EntityId(entity), tag)
}

// Get the part id of a tag.
func (tag Tag) PartId() storage.PartId {
	return storage.PartId(tag)
}
