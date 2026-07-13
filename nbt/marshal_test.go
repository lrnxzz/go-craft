package nbt_test

import (
	"reflect"
	"testing"

	"github.com/lrnxzz/go-craft/nbt"
)

type dimension struct {
	Natural      bool    `nbt:"natural"`
	Height       int32   `nbt:"height"`
	AmbientLight float32 `nbt:"ambient_light"`
	Skip         string  `nbt:"-"`
	Optional     string  `nbt:"optional,omitempty"`
}

type meta struct {
	Version int32 `nbt:"version"`
}

type registry struct {
	meta
	Name    string    `nbt:"name"`
	Palette []string  `nbt:"palette"`
	Blocks  []int64   `nbt:"blocks"`
	Heights []int32   `nbt:"heights,list"`
	Dim     dimension `nbt:"dimension"`
}

func TestMarshalDecodesToExpectedTree(t *testing.T) {
	value := registry{
		meta:    meta{Version: 3},
		Name:    "overworld",
		Palette: []string{"stone", "dirt"},
		Blocks:  []int64{1, 2, 3},
		Heights: []int32{64, 128},
		Dim: dimension{
			Natural:      true,
			Height:       384,
			AmbientLight: 0.5,
			Skip:         "ignored",
		},
	}

	encoded, err := nbt.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := nbt.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}

	want := nbt.Compound{
		"version": nbt.Int(3),
		"name":    nbt.String("overworld"),
		"palette": nbt.List{
			Elem:  nbt.TagString,
			Items: []nbt.Tag{nbt.String("stone"), nbt.String("dirt")},
		},
		"blocks": nbt.LongArray{1, 2, 3},
		"heights": nbt.List{
			Elem:  nbt.TagInt,
			Items: []nbt.Tag{nbt.Int(64), nbt.Int(128)},
		},
		"dimension": nbt.Compound{
			"natural":       nbt.Byte(1),
			"height":        nbt.Int(384),
			"ambient_light": nbt.Float(0.5),
		},
	}

	if !reflect.DeepEqual(decoded, want) {
		t.Errorf("marshal produced:\n got %#v\nwant %#v", decoded, want)
	}
}

func TestMarshalPassesThroughRawTags(t *testing.T) {
	value := map[string]any{
		"raw": nbt.LongArray{7, 8, 9},
	}

	encoded, err := nbt.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := nbt.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}

	if got, ok := nbt.Get[nbt.LongArray](decoded, "raw"); !ok || !reflect.DeepEqual(got, nbt.LongArray{7, 8, 9}) {
		t.Errorf("raw tag passthrough = %#v (ok=%t), want LongArray{7,8,9}", got, ok)
	}
}

func TestMarshalRejectsUnsupportedRoot(t *testing.T) {
	if _, err := nbt.Marshal(42); err == nil {
		t.Error("expected an error marshaling a non-compound root, got nil")
	}
}
