package mesh

import (
	"github.com/go-gl/mathgl/mgl32"
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/viewer/gpu"
)

type Tiles interface {
	Tile(state gocraft.BlockState, face Face) gpu.UV
}

func Chunk(world *gocraft.World, column *gocraft.Column, tiles Tiles) Geometry {
	b := newBuilder()

	baseX := int(column.X) * 16
	baseZ := int(column.Z) * 16
	minY := column.MinY()
	maxY := minY + column.Height()

	for lx := range 16 {
		for lz := range 16 {
			for y := minY; y < maxY; y++ {
				x, z := baseX+lx, baseZ+lz
				state, _ := world.Block(x, y, z)
				if state == 0 {
					continue
				}

				for _, face := range cubeFaces {
					if neighbor, _ := world.Block(x+face.step[0], y+face.step[1], z+face.step[2]); neighbor != 0 {
						continue
					}
					b.quad(mgl32.Vec3{float32(x), float32(y), float32(z)}, face, tiles.Tile(state, face.face))
				}
			}
		}
	}

	return b.geometry()
}
