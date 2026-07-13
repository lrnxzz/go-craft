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
