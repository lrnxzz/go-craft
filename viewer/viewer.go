package viewer

import (
	"image"
	"image/color"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	defaultWidth  = 1280
	defaultHeight = 720
)

var cubeVertexShader = `#version 330 core
layout (location = 0) in vec3 position;
layout (location = 1) in vec2 uv;
uniform mat4 mvp;
out vec2 texCoord;
void main() {
	gl_Position = mvp * vec4(position, 1.0);
	texCoord = uv;
}`

var cubeFragmentShader = `#version 330 core
in vec2 texCoord;
uniform sampler2D atlas;
out vec4 fragColor;
void main() {
	fragColor = texture(atlas, texCoord);
}`

var cubeVertices = []float32{
	-0.5, -0.5, 0.5, 0, 0, 0.5, -0.5, 0.5, 1, 0, 0.5, 0.5, 0.5, 1, 1, -0.5, 0.5, 0.5, 0, 1,
	0.5, -0.5, -0.5, 0, 0, -0.5, -0.5, -0.5, 1, 0, -0.5, 0.5, -0.5, 1, 1, 0.5, 0.5, -0.5, 0, 1,
	0.5, -0.5, 0.5, 0, 0, 0.5, -0.5, -0.5, 1, 0, 0.5, 0.5, -0.5, 1, 1, 0.5, 0.5, 0.5, 0, 1,
	-0.5, -0.5, -0.5, 0, 0, -0.5, -0.5, 0.5, 1, 0, -0.5, 0.5, 0.5, 1, 1, -0.5, 0.5, -0.5, 0, 1,
	-0.5, 0.5, 0.5, 0, 0, 0.5, 0.5, 0.5, 1, 0, 0.5, 0.5, -0.5, 1, 1, -0.5, 0.5, -0.5, 0, 1,
	-0.5, -0.5, -0.5, 0, 0, 0.5, -0.5, -0.5, 1, 0, 0.5, -0.5, 0.5, 1, 1, -0.5, -0.5, 0.5, 0, 1,
}

type Viewer struct {
	window  *Window
	program *Program
	mesh    *Mesh
	texture *Texture
	camera  Camera
}

func New(visible bool) (*Viewer, error) {
	window, err := OpenWindow("gocraft", defaultWidth, defaultHeight, visible)
	if err != nil {
		return nil, err
	}

	program, err := NewProgram(cubeVertexShader, cubeFragmentShader)
	if err != nil {
		window.Close()

		return nil, err
	}

	gl.Enable(gl.DEPTH_TEST)

	viewer := &Viewer{
		window:  window,
		program: program,
		mesh:    NewMesh(cubeVertices, quadIndices(6), Attribute{Location: 0, Size: 3}, Attribute{Location: 1, Size: 2}),
		texture: NewTexture(grassTexture()),
		camera: Camera{
			Position: mgl32.Vec3{0, 0, 3},
			Up:       mgl32.Vec3{0, 1, 0},
			FOV:      45,
			Aspect:   float32(defaultWidth) / float32(defaultHeight),
			Near:     0.1,
			Far:      100,
		},
	}

	return viewer, nil
}

func (v *Viewer) frame(model mgl32.Mat4) {
	gl.ClearColor(0.53, 0.71, 0.92, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	v.program.Use()
	v.program.Mat4("mvp", v.camera.Projection().Mul4(v.camera.View()).Mul4(model))
	v.texture.Bind(0)
	v.program.Int("atlas", 0)
	v.mesh.Draw()
}

func (v *Viewer) Run() {
	defer v.window.Close()

	for !v.window.ShouldClose() {
		spin := float32(glfw.GetTime())
		model := mgl32.HomogRotate3DY(spin * 0.6).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(20)))
		v.frame(model)
		v.window.Present()
	}
}

func (v *Viewer) Screenshot(path string) error {
	defer v.window.Close()

	model := mgl32.HomogRotate3DY(mgl32.DegToRad(35)).Mul4(mgl32.HomogRotate3DX(mgl32.DegToRad(25)))
	v.frame(model)
	gl.Finish()

	return v.window.Capture(path)
}

func quadIndices(quads uint32) []uint32 {
	indices := make([]uint32, 0, quads*6)
	for quad := range quads {
		base := quad * 4
		indices = append(indices, base, base+1, base+2, base, base+2, base+3)
	}

	return indices
}

func grassTexture() image.Image {
	const size = 16
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	light := color.RGBA{106, 170, 80, 255}
	dark := color.RGBA{79, 130, 58, 255}
	for y := range size {
		for x := range size {
			shade := light
			if (x/2+y/2)%2 == 0 {
				shade = dark
			}
			img.SetRGBA(x, y, shade)
		}
	}

	return img
}
