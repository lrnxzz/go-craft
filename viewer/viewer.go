package viewer

import (
	"github.com/go-gl/mathgl/mgl32"
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

const (
	defaultWidth  = 1280
	defaultHeight = 720
	remeshEvery   = 120
)

type Viewer struct {
	window   *gpu.Window
	renderer *Renderer
	camera   Camera
	world    *gocraft.World
}

func New(world *gocraft.World, focus gocraft.Vec3d, visible bool) (*Viewer, error) {
	window, err := gpu.OpenWindow("gocraft", defaultWidth, defaultHeight, visible)
	if err != nil {
		return nil, err
	}

	tileset, err := LoadTileset()
	if err != nil {
		window.Close()

		return nil, err
	}

	renderer, err := NewRenderer(tileset)
	if err != nil {
		window.Close()

		return nil, err
	}
	renderer.Build(world)

	return &Viewer{
		window:   window,
		renderer: renderer,
		camera:   lookAt(focus),
		world:    world,
	}, nil
}

func (v *Viewer) frame() {
	v.window.Clear(0.53, 0.71, 0.92)
	v.renderer.Draw(v.camera)
}

func (v *Viewer) Run() {
	defer v.window.Close()

	for frame := 0; !v.window.ShouldClose(); frame++ {
		if frame%remeshEvery == 0 {
			v.renderer.Build(v.world)
		}
		v.frame()
		v.window.Present()
	}
}

func (v *Viewer) Screenshot(path string) error {
	defer v.window.Close()

	v.frame()

	return v.window.Capture(path)
}

func lookAt(focus gocraft.Vec3d) Camera {
	center := mgl32.Vec3{float32(focus.X), float32(focus.Y), float32(focus.Z)}

	return Camera{
		Position: center.Add(mgl32.Vec3{0, 24, 40}),
		Target:   center,
		Up:       mgl32.Vec3{0, 1, 0},
		FOV:      65,
		Aspect:   float32(defaultWidth) / float32(defaultHeight),
		Near:     0.1,
		Far:      2000,
	}
}
