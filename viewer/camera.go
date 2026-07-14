package viewer

import "github.com/go-gl/mathgl/mgl32"

type Camera struct {
	Position mgl32.Vec3
	Target   mgl32.Vec3
	Up       mgl32.Vec3
	FOV      float32
	Aspect   float32
	Near     float32
	Far      float32
}

func (c Camera) View() mgl32.Mat4 {
	return mgl32.LookAtV(c.Position, c.Target, c.Up)
}

func (c Camera) Projection() mgl32.Mat4 {
	return mgl32.Perspective(mgl32.DegToRad(c.FOV), c.Aspect, c.Near, c.Far)
}
