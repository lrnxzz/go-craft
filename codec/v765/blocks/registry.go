package blocks

import (
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765/assets"
)

//go:generate go run github.com/lrnxzz/go-craft/cli gen blocks 765

var registry = gocraft.LoadRegistry[gocraft.Block](765, assets.Blocks)

var Of = gocraft.Ranged(registry, func(b gocraft.Block) (gocraft.BlockState, gocraft.BlockState) {
	return b.MinState, b.MaxState
})

var Named = gocraft.Keyed(registry, func(b gocraft.Block) gocraft.Identifier {
	return b.Name
})
