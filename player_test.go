package gocraft_test

import (
	"math"
	"testing"

	gocraft "github.com/lrnxzz/go-craft"
)

type lookCase struct {
	yaw   float32
	pitch float32
	want  gocraft.Vec3d
}

func TestPlayerLookDirection(t *testing.T) {
	cases := []lookCase{
		{
			yaw:   0,
			pitch: 0,
			want:  gocraft.Vec3(0, 0, 1),
		},
		{
			yaw:   90,
			pitch: 0,
			want:  gocraft.Vec3(-1, 0, 0),
		},
		{
			yaw:   -90,
			pitch: 0,
			want:  gocraft.Vec3(1, 0, 0),
		},
		{
			yaw:   0,
			pitch: 90,
			want:  gocraft.Vec3(0, -1, 0),
		},
		{
			yaw:   0,
			pitch: -90,
			want:  gocraft.Vec3(0, 1, 0),
		},
	}
	for _, c := range cases {
		player := &gocraft.Player{
			Yaw:   c.yaw,
			Pitch: c.pitch,
		}

		got := player.LookDirection()
		if !got.ApproxEqual(c.want, 1e-9) {
			t.Errorf("LookDirection(yaw %v, pitch %v) = %v, want %v", c.yaw, c.pitch, got, c.want)
		}
	}
}

func TestPlayerEye(t *testing.T) {
	player := &gocraft.Player{
		Position: gocraft.Vec3(1, 64, -3),
	}

	got := player.Eye()
	want := gocraft.Vec3(1, 65.62, -3)
	if !got.ApproxEqual(want, 1e-9) {
		t.Errorf("Eye() = %v, want %v", got, want)
	}
}

func TestPlayerBox(t *testing.T) {
	player := &gocraft.Player{
		Position: gocraft.Vec3(0, 10, 0),
	}

	box := player.Box()
	if box.Min.Y != 10 || box.Max.Y != 11.8 {
		t.Errorf("box spans Y [%v, %v], want [10, 11.8]", box.Min.Y, box.Max.Y)
	}
	if width := box.Max.X - box.Min.X; math.Abs(width-0.6) > 1e-9 {
		t.Errorf("box width = %v, want 0.6", width)
	}
}

func TestPlayerAlive(t *testing.T) {
	player := &gocraft.Player{
		Health: 20,
	}
	if !player.Alive() {
		t.Error("player with full health should be alive")
	}

	player.Health = 0
	if player.Alive() {
		t.Error("player with zero health should be dead")
	}
}
