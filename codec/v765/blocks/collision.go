package blocks

import gocraft "github.com/lrnxzz/go-craft"

var fullCube = []gocraft.AABB{gocraft.Box(gocraft.Vec3d{}, gocraft.Vec3d{X: 1, Y: 1, Z: 1})}

func Collision(state gocraft.BlockState) []gocraft.AABB {
	block, ok := Of(state)
	if !ok || !block.Solid() {
		return nil
	}

	return fullCube
}
