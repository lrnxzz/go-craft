package gocraft_test

import (
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
)

func TestPhysicsLandsOnGround(t *testing.T) {
	column := gocraft.NewChunkColumn(0, 0, -64, 384)
	column.SetBlock(0, 0, 0, 1)

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	player := &gocraft.Player{Position: gocraft.Vec3d{X: 0.5, Y: 5, Z: 0.5}}
	physics := &gocraft.Physics{}

	for range 200 {
		physics.Tick(world, player)
	}

	if !player.OnGround {
		t.Fatal("player should rest on the ground after falling")
	}
	if got := player.Position.Y; got != 1 {
		t.Errorf("resting Y = %v, want 1", got)
	}
}

func TestPhysicsFallsThroughAir(t *testing.T) {
	world := gocraft.NewWorld()
	world.LoadColumn(gocraft.NewChunkColumn(0, 0, -64, 384))

	player := &gocraft.Player{Position: gocraft.Vec3d{X: 0.5, Y: 100, Z: 0.5}}
	physics := &gocraft.Physics{}

	physics.Tick(world, player)

	if player.OnGround {
		t.Error("player over empty space should not be on the ground")
	}
	if player.Position.Y >= 100 {
		t.Errorf("player should have fallen, Y = %v", player.Position.Y)
	}
}
