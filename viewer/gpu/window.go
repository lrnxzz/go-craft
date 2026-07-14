package gpu

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type Window struct {
	handle  *glfw.Window
	width   int
	height  int
	cursorX float64
	cursorY float64
}

func OpenWindow(title string, width, height int, visible bool) (*Window, error) {
	if err := glfw.Init(); err != nil {
		return nil, fmt.Errorf("gpu: glfw init: %w", err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	if !visible {
		glfw.WindowHint(glfw.Visible, glfw.False)
	}

	handle, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		glfw.Terminate()

		return nil, fmt.Errorf("gpu: create window: %w", err)
	}
	handle.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		glfw.Terminate()

		return nil, fmt.Errorf("gpu: gl init: %w", err)
	}
	gl.Enable(gl.DEPTH_TEST)

	return &Window{handle: handle, width: width, height: height}, nil
}

func (w *Window) ShouldClose() bool {
	return w.handle.ShouldClose()
}

func (w *Window) Clear(r, g, b float32) {
	gl.ClearColor(r, g, b, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (w *Window) Present() {
	w.handle.SwapBuffers()
	glfw.PollEvents()
}

func (w *Window) Size() (int, int) {
	return w.width, w.height
}

func (w *Window) Close() {
	glfw.Terminate()
}

func (w *Window) Capture(path string) error {
	gl.Finish()

	pixels := make([]byte, w.width*w.height*4)
	gl.ReadPixels(0, 0, int32(w.width), int32(w.height), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))

	img := image.NewRGBA(image.Rect(0, 0, w.width, w.height))
	stride := w.width * 4
	for y := range w.height {
		copy(img.Pix[y*stride:(y+1)*stride], pixels[(w.height-1-y)*stride:])
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}
