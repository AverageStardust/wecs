package main

import (
	"github.com/averagestardust/wecs/internal/storage"
)

type Tag storage.PartId

var nextTag uint32 = ^uint32(0)

func NewTag() Tag {
	tag := Tag(nextTag)
	nextTag--

	return tag
}

func (tag Tag) Delete(access *Access, entity Entity) (success bool) {
	return access.store.DeletePart(storage.EntityId(entity), tag)
}

func (tag Tag) Has(access *Access, entity Entity) (has bool) {
	return access.store.HasPart(storage.EntityId(entity), tag)
}

func (tag Tag) Add(access *Access, entity Entity) (success bool) {
	return access.store.AddPart(storage.EntityId(entity), tag)
}

func (tag Tag) PartId() storage.PartId {
	return storage.PartId(tag)
}
