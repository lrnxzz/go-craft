package gocraft_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
)

func TestPalettedContainerDecodesSingleValued(t *testing.T) {
	payload := []byte{0}
	payload = gocraft.AppendVar(payload, gocraft.VarInt(42))
	payload = gocraft.AppendVar(payload, gocraft.VarInt(0))

	container := gocraft.NewBlockStates()
	if err := container.Decode(gocraft.NewReader(payload)); err != nil {
		t.Fatal(err)
	}

	for _, index := range []int{0, 100, 4095} {
		if got := container.Get(index); got != 42 {
			t.Errorf("Get(%d) = %d, want 42", index, got)
		}
	}
}

func TestPalettedContainerDecodesIndirect(t *testing.T) {
	payload := []byte{4}

	payload = gocraft.AppendVar(payload, gocraft.VarInt(3))
	for _, value := range []gocraft.VarInt{10, 20, 30} {
		payload = gocraft.AppendVar(payload, value)
	}

	longs := make([]uint64, 256)
	longs[0] = 0<<0 | 1<<4 | 2<<8

	payload = gocraft.AppendVar(payload, gocraft.VarInt(len(longs)))
	for _, long := range longs {
		payload = gocraft.Long(long).Append(payload)
	}

	container := gocraft.NewBlockStates()
	if err := container.Decode(gocraft.NewReader(payload)); err != nil {
		t.Fatal(err)
	}

	want := map[int]gocraft.BlockState{0: 10, 1: 20, 2: 30, 3: 10, 4095: 10}
	for index, state := range want {
		if got := container.Get(index); got != state {
			t.Errorf("Get(%d) = %d, want %d", index, got, state)
		}
	}
}

func TestBlockOfResolvesState(t *testing.T) {
	block, ok := gocraft.BlockOf(2885)
	if !ok {
		t.Fatal("no block for state 2885")
	}
	if block.Name != "oak_stairs" {
		t.Errorf("name = %q, want oak_stairs", block.Name)
	}
}

func TestBlockNamedFindsRange(t *testing.T) {
	block, ok := gocraft.BlockNamed("oak_stairs")
	if !ok {
		t.Fatal("oak_stairs not found")
	}
	if block.MinState != 2874 || block.MaxState != 2953 {
		t.Errorf("range = [%d, %d], want [2874, 2953]", block.MinState, block.MaxState)
	}
}

func TestBlockDecomposesStateProperties(t *testing.T) {
	block, ok := gocraft.BlockNamed("oak_stairs")
	if !ok {
		t.Fatal("oak_stairs not found")
	}

	want := map[string]string{"facing": "north", "half": "top", "shape": "straight", "waterlogged": "true"}
	for name, value := range block.At(block.MinState) {
		if want[name] != value {
			t.Errorf("%s = %q, want %q", name, value, want[name])
		}
	}
}
