package viewer

import "github.com/go-gl/gl/v3.3-core/gl"

type Attribute struct {
	Location uint32
	Size     int32
}

type Mesh struct {
	vao     uint32
	vbo     uint32
	ebo     uint32
	count   int32
	indexed bool
}

func NewMesh(vertices []float32, indices []uint32, layout ...Attribute) *Mesh {
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

	mesh := &Mesh{vao: vao, vbo: vbo, count: int32(len(vertices)) / stride}
	if len(indices) > 0 {
		gl.GenBuffers(1, &mesh.ebo)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)
		mesh.count = int32(len(indices))
		mesh.indexed = true
	}

	return mesh
}

func (m *Mesh) Draw() {
	gl.BindVertexArray(m.vao)
	if m.indexed {
		gl.DrawElements(gl.TRIANGLES, m.count, gl.UNSIGNED_INT, nil)

		return
	}

	gl.DrawArrays(gl.TRIANGLES, 0, m.count)
}
