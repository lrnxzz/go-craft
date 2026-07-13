package nbt

import (
	"bytes"
	"testing"
)

func TestMUTF8RoundTrip(t *testing.T) {
	inputs := []string{
		"",
		"hello",
		"go-craft ⛏",
		"null\x00inside",
		"emoji 😀🔥",
		"grüße café",
	}

	for _, input := range inputs {
		decoded, err := _decodeMUTF8(_encodeMUTF8(nil, input))
		if err != nil {
			t.Errorf("decode(encode(%q)): %v", input, err)
			continue
		}
		if decoded != input {
			t.Errorf("round trip of %q yielded %q", input, decoded)
		}
	}
}

func TestMUTF8NeverEmitsNullByte(t *testing.T) {
	encoded := _encodeMUTF8(nil, "a\x00b")

	if bytes.IndexByte(encoded, 0x00) >= 0 {
		t.Errorf("encoding of an embedded null produced a 0x00 byte: %x", encoded)
	}
}

func TestMUTF8MatchesUTF8ForASCII(t *testing.T) {
	ascii := "plain ascii 123"

	if got := _encodeMUTF8(nil, ascii); !bytes.Equal(got, []byte(ascii)) {
		t.Errorf("ascii encoded to %x, want identical to utf-8 %x", got, ascii)
	}
}

func TestMUTF8EncodesSupplementaryAsSurrogatePair(t *testing.T) {
	if got := len(_encodeMUTF8(nil, "😀")); got != 6 {
		t.Errorf("supplementary char encoded to %d bytes, want 6 (two 3-byte surrogates)", got)
	}
}

func TestMUTF8RejectsTruncatedSequence(t *testing.T) {
	if _, err := _decodeMUTF8([]byte{0xE0, 0x80}); err == nil {
		t.Error("expected an error on a truncated 3-byte sequence, got nil")
	}
}
