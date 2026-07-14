package mesh

import "github.com/go-gl/mathgl/mgl32"

type Face int

const (
	Up Face = iota
	Down
	Side
)

type cubeFace struct {
	face    Face
	step    [3]int
	corners [4]mgl32.Vec3
	shade   float32
}

var cubeFaces = [...]cubeFace{
	{Up, [3]int{0, 1, 0}, [4]mgl32.Vec3{{0, 1, 1}, {1, 1, 1}, {1, 1, 0}, {0, 1, 0}}, 1.0},
	{Down, [3]int{0, -1, 0}, [4]mgl32.Vec3{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}}, 0.5},
	{Side, [3]int{0, 0, 1}, [4]mgl32.Vec3{{1, 0, 1}, {0, 0, 1}, {0, 1, 1}, {1, 1, 1}}, 0.8},
	{Side, [3]int{0, 0, -1}, [4]mgl32.Vec3{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}}, 0.8},
	{Side, [3]int{1, 0, 0}, [4]mgl32.Vec3{{1, 0, 0}, {1, 0, 1}, {1, 1, 1}, {1, 1, 0}}, 0.6},
	{Side, [3]int{-1, 0, 0}, [4]mgl32.Vec3{{0, 0, 1}, {0, 0, 0}, {0, 1, 0}, {0, 1, 1}}, 0.6},
}
