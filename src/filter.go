package main

import (
	"iter"

	"github.com/averagestardust/wecs/internal/storage"
)

type Filter []layer

type layer interface {
	check(archetype *storage.Signature) bool
}

type exactlyLayer storage.Signature
type includeAllLayer storage.Signature
type includeAnyLayer storage.Signature
type excludeLayer storage.Signature

func NewFilter() Filter {
	return Filter{}
}

func (layer exactlyLayer) check(archetype *storage.Signature) bool {
	return archetype.EqualTo(storage.Signature(layer))
}

func (layer includeAllLayer) check(archetype *storage.Signature) bool {
	return archetype.ContainsAll(storage.Signature(layer))
}

func (layer includeAnyLayer) check(archetype *storage.Signature) bool {
	return archetype.ContainsAny(storage.Signature(layer))
}

func (layer excludeLayer) check(archetype *storage.Signature) bool {
	return !archetype.ContainsAny(storage.Signature(layer))
}

func (filter Filter) Exactly(components ...storage.Part) Filter {
	filter = append(filter, exactlyLayer(storage.NewSignature(components)))
	return filter
}

func (filter Filter) IncludeAll(components ...storage.Part) Filter {
	filter = append(filter, includeAllLayer(storage.NewSignature(components)))
	return filter
}

func (filter Filter) IncludeAny(components ...storage.Part) Filter {
	filter = append(filter, includeAnyLayer(storage.NewSignature(components)))
	return filter
}

func (filter Filter) Exclude(components ...storage.Part) Filter {
	filter = append(filter, excludeLayer(storage.NewSignature(components)))
	return filter
}

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
