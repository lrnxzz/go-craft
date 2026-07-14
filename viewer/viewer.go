package viewer

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lrnxzz/go-craft/agent"
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
	bot      *agent.Agent
	yaw      float32
	pitch    float32
}

func New(bot *agent.Agent, visible bool) (*Viewer, error) {
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
	renderer.Build(bot.World())

	spawn := bot.Snapshot()

	return &Viewer{
		window:   window,
		renderer: renderer,
		bot:      bot,
		yaw:      spawn.Yaw,
		pitch:    spawn.Pitch,
		camera: Camera{
			Up:     mgl32.Vec3{0, 1, 0},
			FOV:    70,
			Aspect: float32(defaultWidth) / float32(defaultHeight),
			Near:   0.1,
			Far:    2000,
		},
	}, nil
}

func (v *Viewer) frame() {
	v.follow()
	v.window.Clear(0.53, 0.71, 0.92)
	v.renderer.Draw(v.camera)
}

func (v *Viewer) Run() {
	defer v.window.Close()

	v.window.GrabCursor()
	for frame := 0; !v.window.ShouldClose(); frame++ {
		if frame%remeshEvery == 0 {
			v.renderer.Build(v.bot.World())
		}
		v.control()
		v.frame()
		v.window.Present()
	}
}

func (v *Viewer) Screenshot(path string) error {
	defer v.window.Close()

	v.frame()

	return v.window.Capture(path)
}
