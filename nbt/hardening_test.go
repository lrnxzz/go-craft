package nbt_test

import (
	"reflect"
	"testing"

	"github.com/lrnxzz/go-craft/nbt"
)

func TestEncodeDecodeNamedRoundTrip(t *testing.T) {
	root := nbt.Compound{
		"level": nbt.String("overworld"),
		"seed":  nbt.Long(-4096),
	}

	name, decoded, err := nbt.DecodeNamed(nbt.EncodeNamed("Data", root))
	if err != nil {
		t.Fatal(err)
	}

	if name != "Data" {
		t.Errorf("root name = %q, want Data", name)
	}
	if !reflect.DeepEqual(decoded, root) {
		t.Errorf("named round trip mismatch:\n got %#v\nwant %#v", decoded, root)
	}
}

func TestDecodeRejectsDeeplyNestedInput(t *testing.T) {
	bomb := []byte{byte(nbt.TagCompound), byte(nbt.TagList), 0x00, 0x00}

	for range 1000 {
		bomb = append(bomb, byte(nbt.TagList), 0x00, 0x00, 0x00, 0x01)
	}

	if _, err := nbt.Decode(bomb); err == nil {
		t.Error("expected a nesting-depth error, got nil")
	}
}

func TestDecodeNeverPanicsOnTruncatedInput(t *testing.T) {
	full := nbt.Encode(richTree())

	for cut := range len(full) {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Decode panicked on a %d-byte prefix: %v", cut, r)
				}
			}()

			nbt.Decode(full[:cut])
		}()
	}
}
