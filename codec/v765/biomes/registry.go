package biomes

import (
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765/assets"
	"github.com/lrnxzz/go-craft/lib"
)

//go:generate go run github.com/lrnxzz/go-craft/cli gen biomes 765

var registry = lib.LoadRegistry[gocraft.Biome](765, assets.Biomes)

var Of = lib.Keyed(registry, func(b gocraft.Biome) gocraft.BiomeID {
	return b.ID
})

var Named = lib.Keyed(registry, func(b gocraft.Biome) gocraft.Identifier {
	return b.Name
})
