package mesh

import (
	"sync"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

var vertexPool = sync.Pool{New: func() any { return make([]float32, 0, 8192) }}

type Geometry struct {
	vertices []float32
	indices  []uint32
}

func (g Geometry) Upload() *gpu.Mesh {
	mesh := gpu.NewMesh(g.vertices, g.indices,
		gpu.Attribute{Location: 0, Size: 3},
		gpu.Attribute{Location: 1, Size: 2},
		gpu.Attribute{Location: 2, Size: 1})
	vertexPool.Put(g.vertices[:0])

	return mesh
}

type builder struct {
	vertices []float32
	quads    int
}

func newBuilder() builder {
	return builder{vertices: vertexPool.Get().([]float32)[:0]}
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

func (b *builder) geometry() Geometry {
	return Geometry{vertices: b.vertices, indices: gpu.QuadIndices(b.quads)}
}
