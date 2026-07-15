package viewer

import (
	_ "embed"
	"runtime"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/viewer/gpu"
	"github.com/lrnxzz/go-craft/viewer/mesh"
)

//go:embed assets/shaders/world.vert
var vertexShader string

//go:embed assets/shaders/world.frag
var fragmentShader string

const uploadsPerFrame = 4

type chunkKey struct {
	x int32
	z int32
}

type meshJob struct {
	key    chunkKey
	world  *gocraft.World
	column *gocraft.ChunkColumn
}

type meshResult struct {
	key      chunkKey
	geometry mesh.Geometry
}

type Renderer struct {
	program *gpu.Program
	tileset *Tileset
	chunks  map[chunkKey]*gpu.Mesh
	pending map[chunkKey]bool
	jobs    chan meshJob
	results chan meshResult
}

func NewRenderer(tileset *Tileset) (*Renderer, error) {
	program, err := gpu.NewProgram(vertexShader, fragmentShader)
	if err != nil {
		return nil, err
	}

	renderer := &Renderer{
		program: program,
		tileset: tileset,
		chunks:  map[chunkKey]*gpu.Mesh{},
		pending: map[chunkKey]bool{},
		jobs:    make(chan meshJob, 512),
		results: make(chan meshResult, 512),
	}
	for range runtime.NumCPU() {
		go renderer.mesher()
	}

	return renderer, nil
}

func (r *Renderer) mesher() {
	for job := range r.jobs {
		r.results <- meshResult{key: job.key, geometry: mesh.Chunk(job.world, job.column, r.tileset)}
	}
}

func (r *Renderer) Build(world *gocraft.World) {
	loaded := map[chunkKey]bool{}
	for _, column := range world.Columns() {
		key := chunkKey{column.X, column.Z}
		loaded[key] = true
		if r.chunks[key] != nil || r.pending[key] {
			continue
		}

		select {
		case r.jobs <- meshJob{key: key, world: world, column: column}:
			r.pending[key] = true
		default:
		}
	}

	for key, chunk := range r.chunks {
		if !loaded[key] {
			chunk.Delete()
			delete(r.chunks, key)
		}
	}
}

func (r *Renderer) Collect() {
	for range uploadsPerFrame {
		select {
		case result := <-r.results:
			delete(r.pending, result.key)
			r.chunks[result.key] = result.geometry.Upload()
		default:
			return
		}
	}
}

func (r *Renderer) Flush() {
	for len(r.pending) > 0 {
		result := <-r.results
		delete(r.pending, result.key)
		r.chunks[result.key] = result.geometry.Upload()
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

func (r *Renderer) Close() {
	close(r.jobs)
}
