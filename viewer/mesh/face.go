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

func vec3(x, y, z float32) mgl32.Vec3 {
	return mgl32.Vec3{
		x,
		y,
		z,
	}
}

func offsets(x, y, z int) [3]int {
	return [3]int{
		x,
		y,
		z,
	}
}

var cubeFaces = [...]cubeFace{
	{
		face: Up,
		step: offsets(0, 1, 0),
		corners: [4]mgl32.Vec3{
			vec3(0, 1, 1),
			vec3(1, 1, 1),
			vec3(1, 1, 0),
			vec3(0, 1, 0),
		},
		shade: 1.0,
	},
	{
		face: Down,
		step: offsets(0, -1, 0),
		corners: [4]mgl32.Vec3{
			vec3(0, 0, 0),
			vec3(1, 0, 0),
			vec3(1, 0, 1),
			vec3(0, 0, 1),
		},
		shade: 0.5,
	},
	{
		face: Side,
		step: offsets(0, 0, 1),
		corners: [4]mgl32.Vec3{
			vec3(1, 0, 1),
			vec3(0, 0, 1),
			vec3(0, 1, 1),
			vec3(1, 1, 1),
		},
		shade: 0.8,
	},
	{
		face: Side,
		step: offsets(0, 0, -1),
		corners: [4]mgl32.Vec3{
			vec3(0, 0, 0),
			vec3(1, 0, 0),
			vec3(1, 1, 0),
			vec3(0, 1, 0),
		},
		shade: 0.8,
	},
	{
		face: Side,
		step: offsets(1, 0, 0),
		corners: [4]mgl32.Vec3{
			vec3(1, 0, 0),
			vec3(1, 0, 1),
			vec3(1, 1, 1),
			vec3(1, 1, 0),
		},
		shade: 0.6,
	},
	{
		face: Side,
		step: offsets(-1, 0, 0),
		corners: [4]mgl32.Vec3{
			vec3(0, 0, 1),
			vec3(0, 0, 0),
			vec3(0, 1, 0),
			vec3(0, 1, 1),
		},
		shade: 0.6,
	},
}
