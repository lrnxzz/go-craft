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
	physics := gocraft.NewPhysics(func(state gocraft.BlockState) []gocraft.AABB {
		if state == 0 {
			return nil
		}

		return []gocraft.AABB{gocraft.NewAABB(gocraft.Vec3d{}, gocraft.Vec3d{X: 1, Y: 1, Z: 1})}
	})

	for range 200 {
		physics.Tick(world, player, gocraft.Controls{})
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
	physics := gocraft.NewPhysics(func(state gocraft.BlockState) []gocraft.AABB {
		if state == 0 {
			return nil
		}

		return []gocraft.AABB{gocraft.NewAABB(gocraft.Vec3d{}, gocraft.Vec3d{X: 1, Y: 1, Z: 1})}
	})

	for range 3 {
		physics.Tick(world, player, gocraft.Controls{})
	}

	if player.OnGround {
		t.Error("player over empty space should not be on the ground")
	}
	if player.Position.Y >= 100 {
		t.Errorf("player should have fallen, Y = %v", player.Position.Y)
	}
}

func TestPhysicsWalksForward(t *testing.T) {
	column := gocraft.NewChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, 1)
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	player := &gocraft.Player{Position: gocraft.Vec3d{X: 8, Y: 1, Z: 8}, OnGround: true}
	physics := gocraft.NewPhysics(func(state gocraft.BlockState) []gocraft.AABB {
		if state == 0 {
			return nil
		}

		return []gocraft.AABB{gocraft.NewAABB(gocraft.Vec3d{}, gocraft.Vec3d{X: 1, Y: 1, Z: 1})}
	})

	for range 20 {
		physics.Tick(world, player, gocraft.Controls{Forward: true})
	}

	if player.Position.Z <= 8.5 {
		t.Errorf("walking forward (yaw 0) should advance +Z, got Z = %v", player.Position.Z)
	}
	if !player.OnGround {
		t.Error("player should stay on the ground while walking")
	}
}

func TestPhysicsJumpClearsOneBlock(t *testing.T) {
	column := gocraft.NewChunkColumn(0, 0, -64, 384)
	for x := range 16 {
		for z := range 16 {
			column.SetBlock(x, 0, z, 1)
		}
	}

	world := gocraft.NewWorld()
	world.LoadColumn(column)

	player := &gocraft.Player{Position: gocraft.Vec3d{X: 8, Y: 1, Z: 8}, OnGround: true}
	physics := gocraft.NewPhysics(func(state gocraft.BlockState) []gocraft.AABB {
		if state == 0 {
			return nil
		}

		return []gocraft.AABB{gocraft.NewAABB(gocraft.Vec3d{}, gocraft.Vec3d{X: 1, Y: 1, Z: 1})}
	})

	peak := player.Position.Y
	for range 12 {
		physics.Tick(world, player, gocraft.Controls{Jump: true})
		peak = max(peak, player.Position.Y)
	}

	if peak < 2 {
		t.Errorf("jump should clear a full block, peak Y = %v (want >= 2)", peak)
	}
}
