package main

import (
	"iter"

	"github.com/averagestardust/wecs/internal/storage"
)

// A set of checks or "layers" to filter entities.
type Filter []layer

// A unique check for an entity archetype.
type layer interface {
	check(archetype *storage.Signature) bool
}

type exactlyLayer storage.Signature
type includeExactLayer storage.Signature
type includeAnyLayer storage.Signature
type excludeAnyLayer storage.Signature
type excludeExactLayer storage.Signature

// Create a new set of checks to filter entities.
func NewFilter() Filter {
	return Filter{}
}

func (layer exactlyLayer) check(archetype *storage.Signature) bool {
	return archetype.EqualTo(storage.Signature(layer))
}

func (layer includeExactLayer) check(archetype *storage.Signature) bool {
	return archetype.ContainsAll(storage.Signature(layer))
}

func (layer includeAnyLayer) check(archetype *storage.Signature) bool {
	return archetype.ContainsAny(storage.Signature(layer))
}

func (layer excludeAnyLayer) check(archetype *storage.Signature) bool {
	return !archetype.ContainsAny(storage.Signature(layer))
}

func (layer excludeExactLayer) check(archetype *storage.Signature) bool {
	return !archetype.ContainsAll(storage.Signature(layer))
}

// Add a check for entities that all the components exactly match some components.
func (filter Filter) Exactly(components ...storage.Part) Filter {
	filter = append(filter, exactlyLayer(storage.NewSignature(components)))
	return filter
}

// Add a check for entities that have all of some components.
func (filter Filter) IncludeExact(components ...storage.Part) Filter {
	filter = append(filter, includeExactLayer(storage.NewSignature(components)))
	return filter
}

// Add a check for entities that have at least one of some components.
func (filter Filter) IncludeAny(components ...storage.Part) Filter {
	filter = append(filter, includeAnyLayer(storage.NewSignature(components)))
	return filter
}

// Add a check for entities that don't have any of some components.
func (filter Filter) ExcludeAny(components ...storage.Part) Filter {
	filter = append(filter, excludeAnyLayer(storage.NewSignature(components)))
	return filter
}

// Add a check for entities that don't have all of some components.
func (filter Filter) ExcludeExact(components ...storage.Part) Filter {
	filter = append(filter, excludeExactLayer(storage.NewSignature(components)))
	return filter
}

// Filter through the archetypes on a storage, and return the an iterator of matching entities.
func (layers Filter) filter(store *storage.Store) iter.Seq[*storage.Page] {
	return func(yield func(page *storage.Page) bool) {
	pageLoop:
		for archetypeId, page := range store.Pages {
			// check layers
			for _, layer := range layers {
				archetype := &store.Archetypes[archetypeId]
				if !layer.check(archetype) {
					continue pageLoop
				}
			}

			// yield matching page
			if !yield(page) {
				return
			}
		}

	}
}
