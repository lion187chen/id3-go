// Copyright 2013 Michael Yang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package encodedbytes

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSynch(t *testing.T) {
	synch := []byte{0x44, 0x7a, 0x70, 0x04}
	const synchResult = 144619524

	if result, err := SynchInt(synch); result != synchResult {
		t.Errorf("encodedbytes.SynchInt(%v) = %d with error %v, want %d", synch, result, err, synchResult)
	}
	if result := SynchBytes(synchResult); !bytes.Equal(result, synch) {
		t.Errorf("encodedbytes.SynchBytes(%d) = %v, want %v", synchResult, result, synch)
	}
}

func TestNorm(t *testing.T) {
	norm := []byte{0x0b, 0x95, 0xae, 0xb4}
	const normResult = 194358964

	if result, err := NormInt(norm); result != normResult {
		t.Errorf("encodedbytes.NormInt(%v) = %d with error %v, want %d", norm, result, err, normResult)
	}
	if result := NormBytes(normResult); !bytes.Equal(result, norm) {
		t.Errorf("encodedbytes.NormBytes(%d) = %v, want %v", normResult, result, norm)
	}
}

func TestIndexes(t *testing.T) {
	// Encodings, in index order
	encodings := []string{
		"ISO-8859-1", "UTF-16", "UTF-16BE", "UTF-8",
	}
	for i, e := range encodings {
		idx := IndexForEncoding(e)
		assert.Equal(t, byte(i), idx)

		name := EncodingForIndex(byte(i))
		assert.Equal(t, e, name)
	}
}

// Verify that ISO-8859-1 can be decoded and encoded.
func TestEncodeDecode(t *testing.T) {
	// hêllo wørld (e-circumflex in hello, o-slash in world)
	sampleISO_8859_1 := []byte{0x68, 0xea, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0xf8, 0x72, 0x6c, 0x64}
	expectedUTF8 := "hêllo wørld"

	idx := IndexForEncoding("ISO-8859-1")
	decoded, err := Decoders[idx].ConvertString(string(sampleISO_8859_1))
	require.NoError(t, err)
	assert.Equal(t, expectedUTF8, decoded)

	// Try round-tripping it, and compare with original.
	encoded, err := Encoders[idx].ConvertString(decoded)
	require.NoError(t, err)
	assert.Equal(t, string(sampleISO_8859_1), encoded)
}
