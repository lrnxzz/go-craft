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

type Renderer struct {
	program *gpu.Program
	tileset *Tileset
	chunks  []*gpu.Mesh
}

func NewRenderer(tileset *Tileset) (*Renderer, error) {
	program, err := gpu.NewProgram(vertexShader, fragmentShader)
	if err != nil {
		return nil, err
	}

	return &Renderer{program: program, tileset: tileset}, nil
}

func (r *Renderer) Build(world *gocraft.World) {
	r.chunks = r.chunks[:0]
	for _, column := range world.Columns() {
		r.chunks = append(r.chunks, mesh.Chunk(world, column, r.tileset))
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
