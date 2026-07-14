package viewer

import "github.com/go-gl/gl/v3.3-core/gl"

type Attribute struct {
	Location uint32
	Size     int32
}

type Mesh struct {
	vao      uint32
	vbo      uint32
	vertices int32
}

func NewMesh(vertices []float32, layout ...Attribute) *Mesh {
	var stride int32
	for _, attr := range layout {
		stride += attr.Size
	}

	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	var offset int32
	for _, attr := range layout {
		gl.EnableVertexAttribArray(attr.Location)
		gl.VertexAttribPointerWithOffset(attr.Location, attr.Size, gl.FLOAT, false, stride*4, uintptr(offset*4))
		offset += attr.Size
	}

	return &Mesh{vao: vao, vbo: vbo, vertices: int32(len(vertices)) / stride}
}

func (m *Mesh) Draw() {
	gl.BindVertexArray(m.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, m.vertices)
}
