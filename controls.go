package gocraft

import "math"

const (
	walkSpeed    = 0.21
	sprintSpeed  = 0.28
	jumpVelocity = 0.42
)

type Controls struct {
	Forward bool
	Back    bool
	Left    bool
	Right   bool
	Jump    bool
	Sprint  bool
}

func (c Controls) heading(yaw float32) Vec3d {
	var forward, strafe float64
	if c.Forward {
		forward++
	}
	if c.Back {
		forward--
	}
	if c.Left {
		strafe++
	}
	if c.Right {
		strafe--
	}
	if forward == 0 && strafe == 0 {
		return Vec3d{}
	}

	rad := float64(yaw) * math.Pi / 180
	f, g := math.Sin(rad), math.Cos(rad)

	return Vec3(strafe*g-forward*f, 0, forward*g+strafe*f).Normalize()
}
