package gocraft

import (
	"fmt"
	"math"
)

type Vec3d struct {
	X float64
	Y float64
	Z float64
}

func (v Vec3d) Add(o Vec3d) Vec3d {
	return Vec3d{v.X + o.X, v.Y + o.Y, v.Z + o.Z}
}

func (v Vec3d) Sub(o Vec3d) Vec3d {
	return Vec3d{v.X - o.X, v.Y - o.Y, v.Z - o.Z}
}

func (v Vec3d) Mul(o Vec3d) Vec3d {
	return Vec3d{v.X * o.X, v.Y * o.Y, v.Z * o.Z}
}

func (v Vec3d) Scale(s float64) Vec3d {
	return Vec3d{v.X * s, v.Y * s, v.Z * s}
}

func (v Vec3d) Neg() Vec3d {
	return Vec3d{-v.X, -v.Y, -v.Z}
}

func (v Vec3d) Offset(dx, dy, dz float64) Vec3d {
	return Vec3d{v.X + dx, v.Y + dy, v.Z + dz}
}

func (v Vec3d) Dot(o Vec3d) float64 {
	return v.X*o.X + v.Y*o.Y + v.Z*o.Z
}

func (v Vec3d) Cross(o Vec3d) Vec3d {
	return Vec3d{
		X: v.Y*o.Z - v.Z*o.Y,
		Y: v.Z*o.X - v.X*o.Z,
		Z: v.X*o.Y - v.Y*o.X,
	}
}

func (v Vec3d) LengthSquared() float64 {
	return v.Dot(v)
}

func (v Vec3d) Length() float64 {
	return math.Sqrt(v.LengthSquared())
}

func (v Vec3d) HorizontalLength() float64 {
	return math.Hypot(v.X, v.Z)
}

func (v Vec3d) DistanceSquared(o Vec3d) float64 {
	return o.Sub(v).LengthSquared()
}

func (v Vec3d) Distance(o Vec3d) float64 {
	return o.Sub(v).Length()
}

func (v Vec3d) Normalize() Vec3d {
	length := v.Length()
	if length == 0 {
		return Vec3d{}
	}

	return v.Scale(1 / length)
}

func (v Vec3d) Lerp(o Vec3d, t float64) Vec3d {
	return Vec3d{
		X: v.X + (o.X-v.X)*t,
		Y: v.Y + (o.Y-v.Y)*t,
		Z: v.Z + (o.Z-v.Z)*t,
	}
}

func (v Vec3d) Floor() Position {
	return Position{
		X: int(math.Floor(v.X)),
		Y: int(math.Floor(v.Y)),
		Z: int(math.Floor(v.Z)),
	}
}

func (v Vec3d) ApproxEqual(o Vec3d, epsilon float64) bool {
	return math.Abs(v.X-o.X) <= epsilon &&
		math.Abs(v.Y-o.Y) <= epsilon &&
		math.Abs(v.Z-o.Z) <= epsilon
}

func (v Vec3d) String() string {
	return fmt.Sprintf("(%.3f, %.3f, %.3f)", v.X, v.Y, v.Z)
}
