package main

import (
	"encoding/binary"
	"hash/crc64"
	"reflect"

	"github.com/averagestardust/wecs/internal/common"
	"github.com/averagestardust/wecs/internal/storage"
)

type Resource[T any] storage.ResourceId

var resourceTypes []reflect.Type

func NewResource[T any]() Resource[T] {
	var t T
	typ := reflect.TypeOf(t)

	resource := Resource[T](len(resourceTypes))
	resourceTypes = append(resourceTypes, typ)

	return resource
}

func (resource Resource[T]) Delete(store *storage.Store) (success bool) {
	if !resource.Has(store) {
		return false
	}

	delete(store.Resources, storage.ResourceId(resource))
	return true
}

func (resourceId Resource[T]) Get(store *storage.Store) *T {
	resource, exists := store.Resources[storage.ResourceId(resourceId)]
	if !exists {
		return nil
	}

	return resource.(*T)
}

func (resource Resource[T]) Has(store *storage.Store) bool {
	_, exists := store.Resources[storage.ResourceId(resource)]
	return exists
}

func (resource Resource[T]) Add(store *storage.Store) (success bool) {
	if resource.Has(store) {
		return false
	}

	var t T
	store.Resources[storage.ResourceId(resource)] = t
	return true
}

func hashUsedResources(store *storage.Store) uint64 {
	hash := crc64.New(common.Crc64ISOTable)
	uint32Bytes := make([]byte, 4)

	for resourceId := range store.Resources {
		binary.LittleEndian.PutUint32(uint32Bytes, uint32(resourceId))
		hash.Write(uint32Bytes)

		if int(resourceId) >= len(resourceTypes) {
			continue
		}

		resourceType := resourceTypes[resourceId]
		hash.Write([]byte(resourceType.Name()))
	}

	return hash.Sum64()
}
