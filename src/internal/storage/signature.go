package storage

import (
	"encoding/binary"
	"hash/crc64"
	"slices"

	"github.com/averagestardust/wecs/internal/common"
)

type Signature []Part

func NewSignature(parts []Part) Signature {
	PartSet := map[Part]struct{}{}
	for _, Part := range parts {
		PartSet[Part] = struct{}{}
	}

	uniqueParts := []Part{}
	for Part := range PartSet {
		uniqueParts = append(uniqueParts, Part)
	}

	// sort for quick comparison
	slices.SortFunc(uniqueParts, cmpPart)

	return uniqueParts
}

func (signature Signature) Add(newPart Part) Signature {
	parts := slices.Clone(signature)
	parts = append(parts, newPart)

	return NewSignature(parts)
}

func (signature Signature) Delete(removedPart Part) Signature {
	parts := []Part{}

	for _, Part := range signature {
		if Part != removedPart {
			parts = append(parts, Part)
		}
	}

	return NewSignature(parts)
}

func (signature Signature) EqualTo(other Signature) bool {
	if len(signature) != len(other) {
		return false
	}

	// signatures are sorted, thus equal sets have the same elements in the same order
	for i, Component := range signature {
		if Component != other[i] {
			return false
		}
	}

	return true
}

func (signature Signature) ContainsAll(other Signature) bool {
	i := 0
	for _, component := range other {
		// Signatures are in a sorted order, thus we only need to search areas after the previous match
		n, found := slices.BinarySearchFunc(signature[i:], component, cmpPart)
		i += n + 1

		if !found {
			return false
		}
	}

	return true
}

func (signature Signature) ContainsAny(other Signature) bool {
	for _, component := range other {
		if _, found := slices.BinarySearchFunc(signature, component, cmpPart); found {
			return true
		}
	}

	return false
}

func (signature Signature) ContainsSingle(Part Part) bool {
	_, found := slices.BinarySearchFunc(signature, Part, cmpPart)
	return found
}

func (signature Signature) hash() uint64 {
	hash := crc64.New(common.Crc64ISOTable)
	var bytes [4]byte

	for _, Part := range signature {
		binary.LittleEndian.PutUint32(bytes[:], uint32(Part.PartId()))
		hash.Write(bytes[:])
	}

	return hash.Sum64()
}
