package viewer

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

const (
	eyeHeight   = 1.62
	sensitivity = 0.15
)

func (v *Viewer) control() {
	v.bot.SetControls(gocraft.Controls{
		Forward: v.window.Pressed(gpu.KeyW),
		Back:    v.window.Pressed(gpu.KeyS),
		Left:    v.window.Pressed(gpu.KeyA),
		Right:   v.window.Pressed(gpu.KeyD),
		Jump:    v.window.Pressed(gpu.KeySpace),
		Sprint:  v.window.Pressed(gpu.KeyCtrl),
	})

	dx, dy := v.window.CursorDelta()
	v.yaw += float32(dx) * sensitivity
	v.pitch = clamp(v.pitch+float32(dy)*sensitivity, -89, 89)
	v.bot.Look(v.yaw, v.pitch)
}

func (v *Viewer) follow() {
	position := v.bot.Snapshot().Position
	eye := mgl32.Vec3{float32(position.X), float32(position.Y) + eyeHeight, float32(position.Z)}

	v.camera.Position = eye
	v.camera.Target = eye.Add(direction(v.yaw, v.pitch))
}

func direction(yaw, pitch float32) mgl32.Vec3 {
	y := float64(mgl32.DegToRad(yaw))
	p := float64(mgl32.DegToRad(pitch))

	return mgl32.Vec3{
		float32(-math.Sin(y) * math.Cos(p)),
		float32(-math.Sin(p)),
		float32(math.Cos(y) * math.Cos(p)),
	}
}

func clamp(value, low, high float32) float32 {
	return min(max(value, low), high)
}
