package blocks

import (
	"math"

	gocraft "github.com/lrnxzz/go-craft"
)

//go:generate go run github.com/lrnxzz/go-craft/cli gen materials 765

type toolSpeeds = map[gocraft.ItemID]float64

func DigDamage(state gocraft.BlockState, held gocraft.ItemID) (float64, bool) {
	block, ok := Of(state)
	if !ok || !block.Diggable || block.Hardness < 0 {
		return 0, false
	}
	if block.Hardness == 0 {
		return 1, true
	}

	speed := 1.0
	tools, ok := materials[block.Material]
	if ok {
		multiplier, effective := tools[held]
		if effective {
			speed = multiplier
		}
	}

	damage := speed / block.Hardness
	if block.HarvestTools == nil || block.HarvestTools[held] {
		damage /= 30
	} else {
		damage /= 100
	}

	return damage, true
}

func BreakTicks(state gocraft.BlockState, held gocraft.ItemID) (int, bool) {
	damage, ok := DigDamage(state, held)
	if !ok {
		return 0, false
	}
	if damage >= 1 {
		return 0, true
	}

	return int(math.Ceil(1 / damage)), true
}
