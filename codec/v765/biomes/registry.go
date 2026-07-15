package biomes

import (
	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765/assets"
)

var registry = gocraft.LoadRegistry[gocraft.Biome](765, assets.Biomes)

var Of = gocraft.Keyed(registry, func(b gocraft.Biome) gocraft.BiomeID {
	return b.ID
})

var Named = gocraft.Keyed(registry, func(b gocraft.Biome) gocraft.Identifier {
	return b.Name
})
