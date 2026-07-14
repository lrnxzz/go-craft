package viewer

import (
	_ "embed"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/viewer/gpu"
	"github.com/lrnxzz/go-craft/viewer/mesh"
)

//go:embed assets/shaders/world.vert
var vertexShader string

//go:embed assets/shaders/world.frag
var fragmentShader string

type chunkKey struct {
	x int32
	z int32
}

type Renderer struct {
	program *gpu.Program
	tileset *Tileset
	chunks  map[chunkKey]*gpu.Mesh
}

func NewRenderer(tileset *Tileset) (*Renderer, error) {
	program, err := gpu.NewProgram(vertexShader, fragmentShader)
	if err != nil {
		return nil, err
	}

	return &Renderer{program: program, tileset: tileset, chunks: map[chunkKey]*gpu.Mesh{}}, nil
}

func (r *Renderer) Build(world *gocraft.World) {
	loaded := map[chunkKey]bool{}
	for _, column := range world.Columns() {
		key := chunkKey{column.X, column.Z}
		loaded[key] = true
		if _, meshed := r.chunks[key]; !meshed {
			r.chunks[key] = mesh.Chunk(world, column, r.tileset)
		}
	}

	for key, chunk := range r.chunks {
		if !loaded[key] {
			chunk.Delete()
			delete(r.chunks, key)
		}
	}
}

func (r *Renderer) Draw(camera Camera) {
	r.program.Use()
	r.program.Mat4("viewProjection", camera.ViewProjection())
	r.tileset.Atlas().Bind(0)
	r.program.Int("atlas", 0)
	for _, chunk := range r.chunks {
		chunk.Draw()
	}
}
