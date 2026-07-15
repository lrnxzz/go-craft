package viewer

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

const (
	eyeHeight       = 1.62
	sensitivity     = 0.15
	tickSeconds     = 0.05
	doubleTapWindow = 0.3
)

func (v *Viewer) control() {
	forward := v.window.Pressed(gpu.KeyW)

	now := v.window.Time()
	if forward && !v.wHeld {
		if now-v.lastW < doubleTapWindow {
			v.sprinting = true
		}
		v.lastW = now
	}
	if !forward {
		v.sprinting = false
	}
	v.wHeld = forward

	v.bot.SetControls(gocraft.Controls{
		Forward: forward,
		Back:    v.window.Pressed(gpu.KeyS),
		Left:    v.window.Pressed(gpu.KeyA),
		Right:   v.window.Pressed(gpu.KeyD),
		Jump:    v.window.Pressed(gpu.KeySpace),
		Sprint:  v.sprinting || v.window.Pressed(gpu.KeyCtrl),
	})

	dx, dy := v.window.CursorDelta()
	v.yaw += float32(dx) * sensitivity
	v.pitch = clamp(v.pitch+float32(dy)*sensitivity, -89, 89)
	v.bot.Look(v.yaw, v.pitch)
}

func (v *Viewer) follow() {
	snapshot := v.bot.Snapshot()
	if snapshot.Tick != v.lastTick {
		v.lastTick = snapshot.Tick
		v.from = v.to
		v.to = eyeOf(snapshot.Position)
		v.since = v.window.Time()
	}

	alpha := float32(min((v.window.Time()-v.since)/tickSeconds, 1))
	eye := v.from.Add(v.to.Sub(v.from).Mul(alpha))

	v.camera.Position = eye
	v.camera.Target = eye.Add(direction(v.yaw, v.pitch))
}

func eyeOf(position gocraft.Vec3d) mgl32.Vec3 {
	return mgl32.Vec3{float32(position.X), float32(position.Y) + eyeHeight, float32(position.Z)}
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
