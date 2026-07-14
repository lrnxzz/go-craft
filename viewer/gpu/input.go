package gpu

import "github.com/go-gl/glfw/v3.3/glfw"

type Key int

const (
	KeyW     Key = Key(glfw.KeyW)
	KeyA     Key = Key(glfw.KeyA)
	KeyS     Key = Key(glfw.KeyS)
	KeyD     Key = Key(glfw.KeyD)
	KeySpace Key = Key(glfw.KeySpace)
	KeyShift Key = Key(glfw.KeyLeftShift)
	KeyCtrl  Key = Key(glfw.KeyLeftControl)
)

func (w *Window) Pressed(key Key) bool {
	return w.handle.GetKey(glfw.Key(key)) == glfw.Press
}

func (w *Window) GrabCursor() {
	w.handle.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	w.cursorX, w.cursorY = w.handle.GetCursorPos()
}

func (w *Window) CursorDelta() (float64, float64) {
	x, y := w.handle.GetCursorPos()
	dx, dy := x-w.cursorX, y-w.cursorY
	w.cursorX, w.cursorY = x, y

	return dx, dy
}
