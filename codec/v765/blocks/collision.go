package blocks

import gocraft "github.com/lrnxzz/go-craft"

var fullCube = []gocraft.AABB{
	gocraft.Box(gocraft.Vec3(0, 0, 0), gocraft.Vec3(1, 1, 1)),
}

func Solid(state gocraft.BlockState) bool {
	block, ok := Of(state)
	if !ok {
		return false
	}

	return block.Solid()
}

func Collision(state gocraft.BlockState) []gocraft.AABB {
	if !Solid(state) {
		return nil
	}

	return fullCube
}
