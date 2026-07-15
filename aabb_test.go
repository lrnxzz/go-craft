package gocraft_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
)

func TestAABBIntersects(t *testing.T) {
	block := gocraft.Box(gocraft.Vec3(0, 0, 0), gocraft.Vec3(1, 1, 1))
	overlapping := gocraft.Box(gocraft.Vec3(0.5, 0.5, 0.5), gocraft.Vec3(1.5, 1.5, 1.5))
	touching := gocraft.Box(gocraft.Vec3(1, 0, 0), gocraft.Vec3(2, 1, 1))

	if !block.Intersects(overlapping) {
		t.Error("overlapping boxes should intersect")
	}
	if block.Intersects(touching) {
		t.Error("boxes touching only on a face should not intersect")
	}
}

func TestAABBClampYStopsFall(t *testing.T) {
	ground := gocraft.Box(gocraft.Vec3(0, 0, 0), gocraft.Vec3(1, 1, 1))
	player := gocraft.BoxAround(gocraft.Vec3(0.5, 1.5, 0.5), 0.6, 1.8)

	if got := ground.ClampY(player, -1); got != -0.5 {
		t.Errorf("ClampY = %v, want -0.5", got)
	}
}

func TestAABBStretchSweepsAlongSign(t *testing.T) {
	swept := gocraft.Box(gocraft.Vec3(0, 0, 0), gocraft.Vec3(1, 1, 1)).Stretch(-2, 0, 3)

	if swept.Min.X != -2 || swept.Max.X != 1 {
		t.Errorf("X = [%v, %v], want [-2, 1]", swept.Min.X, swept.Max.X)
	}
	if swept.Min.Z != 0 || swept.Max.Z != 4 {
		t.Errorf("Z = [%v, %v], want [0, 4]", swept.Min.Z, swept.Max.Z)
	}
}
