package storage

import (
	"hash/crc64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSignature(t *testing.T) {
	tag0 := partMock(^uint32(0))
	tag2 := partMock(^uint32(2))
	component1 := partMock(1)
	component5 := partMock(5)

	assert.EqualValues(t,
		Signature{component1, component5, tag2, tag0},
		NewSignature([]Part{tag0, component5, component5, tag2, component1, tag2}))

	assert.EqualValues(t,
		Signature{component5, tag2, tag0},
		NewSignature([]Part{tag0, tag2, component5, tag2}))
}

func TestSignatureAdd(t *testing.T) {
	tag1 := partMock(^uint32(1))
	tag4 := partMock(^uint32(4))
	component0 := partMock(0)
	component7 := partMock(7)

	sig := Signature{component0, tag1}

	assert.EqualValues(t, Signature{component0, tag1}, sig.Add(component0))
	assert.EqualValues(t, Signature{component0, component7, tag1}, sig.Add(component7))
	assert.EqualValues(t, Signature{component0, tag4, tag1}, sig.Add(tag4))
}

func TestSignatureDelete(t *testing.T) {
	tag1 := partMock(^uint32(1))
	tag4 := partMock(^uint32(4))
	component0 := partMock(0)
	component7 := partMock(7)

	sig := Signature{component0, tag4, tag1}

	assert.EqualValues(t, Signature{component0, tag4, tag1}, sig.Delete(component7))
	assert.EqualValues(t, Signature{tag4, tag1}, sig.Delete(component0))
	assert.EqualValues(t, Signature{component0, tag1}, sig.Delete(tag4))
}

func TestSignatureEqualTo(t *testing.T) {
	tag1 := partMock(^uint32(1))
	tag4 := partMock(^uint32(4))
	component0 := partMock(0)
	component7 := partMock(7)

	sig := Signature{component0, tag1}

	assert.True(t, sig.EqualTo(sig))
	assert.True(t, sig.EqualTo(Signature{component0, tag1}))

	assert.False(t, sig.EqualTo(Signature{component0, tag4, tag1}))
	assert.False(t, sig.EqualTo(Signature{component0, component7, tag1}))
	assert.False(t, sig.EqualTo(Signature{tag1}))
}

func TestSignatureContainsAll(t *testing.T) {
	tag1 := partMock(^uint32(1))
	tag2 := partMock(^uint32(2))
	tag4 := partMock(^uint32(4))
	component0 := partMock(0)
	component7 := partMock(7)

	sig := Signature{component0, tag4, tag1}

	assert.True(t, sig.ContainsAll(sig))
	assert.True(t, sig.ContainsAll(Signature{tag4}))
	assert.True(t, sig.ContainsAll(Signature{component0, tag1}))

	assert.False(t, sig.ContainsAll(Signature{component7, tag1}))
	assert.False(t, sig.ContainsAll(Signature{component0, tag2, tag1}))
}

func TestSignatureContainsAny(t *testing.T) {
	tag1 := partMock(^uint32(1))
	tag2 := partMock(^uint32(2))
	tag4 := partMock(^uint32(4))
	component0 := partMock(0)
	component3 := partMock(3)
	component7 := partMock(7)

	sig := Signature{component0, tag4, tag1}

	assert.True(t, sig.ContainsAny(sig))
	assert.True(t, sig.ContainsAny(Signature{component0}))
	assert.True(t, sig.ContainsAny(Signature{tag4, component3}))
	assert.True(t, sig.ContainsAny(Signature{tag1, tag4, component0, component7}))

	assert.False(t, sig.ContainsAny(Signature{component3, component7}))
	assert.False(t, sig.ContainsAny(Signature{tag2}))
}

func TestSignatureContainsSingle(t *testing.T) {
	tag1 := partMock(^uint32(1))
	tag2 := partMock(^uint32(2))
	tag4 := partMock(^uint32(4))
	component3 := partMock(3)
	component7 := partMock(7)

	sig := Signature{component3, tag4, tag1}

	assert.True(t, sig.ContainsSingle(tag1))
	assert.True(t, sig.ContainsSingle(component3))

	assert.False(t, sig.ContainsSingle(tag2))
	assert.False(t, sig.ContainsSingle(component7))
}

func TestSignatureHash(t *testing.T) {
	tag0 := partMock(^uint32(1))
	component1 := partMock(1)
	component5 := partMock(5)

	sig := Signature{component1, component5, tag0}

	isoTable := crc64.MakeTable(crc64.ISO)
	assert.EqualValues(t,
		crc64.Checksum([]byte{1, 0, 0, 0, 5, 0, 0, 0, 254, 255, 255, 255}, isoTable),
		sig.hash())
}
