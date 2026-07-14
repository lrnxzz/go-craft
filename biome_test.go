package gocraft_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
)

func TestBiomeOfResolvesID(t *testing.T) {
	info, ok := gocraft.BiomeOf(0)
	if !ok {
		t.Fatal("no biome for id 0")
	}
	if info.Name != "badlands" {
		t.Errorf("name = %q, want badlands", info.Name)
	}
}

func TestBiomeNamedResolves(t *testing.T) {
	info, ok := gocraft.BiomeNamed("plains")
	if !ok {
		t.Fatal("plains not found")
	}
	if info.Dimension != "overworld" {
		t.Errorf("dimension = %q, want overworld", info.Dimension)
	}
}
