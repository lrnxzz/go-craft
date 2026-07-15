package gocraft_test

import (
	"reflect"
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/nbt"
)

func TestNBTFieldAdvancesReader(t *testing.T) {
	original := gocraft.NBT{
		"text": nbt.String("hello"),
		"n":    nbt.Int(7),
	}

	payload := gocraft.AppendAll(nil, original, gocraft.VarInt(42))

	var (
		got     gocraft.NBT
		trailer gocraft.VarInt
	)
	if err := gocraft.Unmarshal(payload, &got, &trailer); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, original) {
		t.Errorf("nbt round trip got %#v, want %#v", got, original)
	}
	if trailer != 42 {
		t.Errorf("trailer = %d, want 42 (nbt decode did not advance the reader exactly)", trailer)
	}
}

func TestNBTDecodesTagEndAsAbsent(t *testing.T) {
	payload := gocraft.AppendAll(nil, gocraft.NBT(nil), gocraft.VarInt(9))

	var (
		decoded gocraft.NBT
		trailer gocraft.VarInt
	)
	if err := gocraft.Unmarshal(payload, &decoded, &trailer); err != nil {
		t.Fatal(err)
	}

	if decoded != nil {
		t.Errorf("decoded = %v, want nil", decoded)
	}
	if trailer != 9 {
		t.Errorf("trailer = %d, want 9 (end tag must consume exactly one byte)", trailer)
	}
}
