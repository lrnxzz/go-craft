package viewer

import "github.com/go-gl/gl/v3.3-core/gl"

const (
	defaultWidth  = 1280
	defaultHeight = 720
)

var triangleVertex = `#version 330 core
layout (location = 0) in vec3 position;
layout (location = 1) in vec3 color;
out vec3 vertexColor;
void main() {
	gl_Position = vec4(position, 1.0);
	vertexColor = color;
}`

var triangleFragment = `#version 330 core
in vec3 vertexColor;
out vec4 fragColor;
void main() {
	fragColor = vec4(vertexColor, 1.0);
}`

var triangle = []float32{
	+0.0, +0.6, 0.0, 0.90, 0.32, 0.26,
	-0.6, -0.5, 0.0, 0.36, 0.72, 0.38,
	+0.6, -0.5, 0.0, 0.30, 0.48, 0.86,
}

type Viewer struct {
	window  *Window
	program *Program
	mesh    *Mesh
}

func New(visible bool) (*Viewer, error) {
	window, err := OpenWindow("gocraft", defaultWidth, defaultHeight, visible)
	if err != nil {
		return nil, err
	}

	program, err := NewProgram(triangleVertex, triangleFragment)
	if err != nil {
		window.Close()

		return nil, err
	}

	mesh := NewMesh(triangle, Attribute{Location: 0, Size: 3}, Attribute{Location: 1, Size: 3})

	return &Viewer{window: window, program: program, mesh: mesh}, nil
}

func (v *Viewer) frame() {
	gl.ClearColor(0.11, 0.13, 0.18, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	v.program.Use()
	v.mesh.Draw()
}

func (v *Viewer) Run() {
	defer v.window.Close()

	for !v.window.ShouldClose() {
		v.frame()
		v.window.Present()
	}
}

func (v *Viewer) Screenshot(path string) error {
	defer v.window.Close()

	v.frame()
	gl.Finish()

	return v.window.Capture(path)
}
