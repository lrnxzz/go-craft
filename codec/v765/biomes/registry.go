package biomes

import (
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765/assets"
)

var registry = gocraft.LoadRegistry[gocraft.BiomeInfo](765, assets.Biomes)

var Of = gocraft.Keyed(registry, func(b gocraft.BiomeInfo) gocraft.Biome {
	return b.ID
})

var Named = gocraft.Keyed(registry, func(b gocraft.BiomeInfo) gocraft.Identifier {
	return b.Name
})
