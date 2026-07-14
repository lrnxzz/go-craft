package mesh

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

type builder struct {
	vertices []float32
	quads    int
}

func (b *builder) quad(origin mgl32.Vec3, face cubeFace, uv gpu.UV) {
	texels := [4]mgl32.Vec2{{uv.U0, uv.V1}, {uv.U1, uv.V1}, {uv.U1, uv.V0}, {uv.U0, uv.V0}}
	for i, corner := range face.corners {
		position := origin.Add(corner)
		b.vertices = append(b.vertices,
			position.X(), position.Y(), position.Z(),
			texels[i].X(), texels[i].Y(),
			face.shade)
	}
	b.quads++
}

func (b *builder) upload() *gpu.Mesh {
	return gpu.NewMesh(b.vertices, gpu.QuadIndices(b.quads),
		gpu.Attribute{Location: 0, Size: 3},
		gpu.Attribute{Location: 1, Size: 2},
		gpu.Attribute{Location: 2, Size: 1})
}
