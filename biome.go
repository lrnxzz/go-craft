package gocraft

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

type Biome int32

var biomePalette = paletteKind{entries: 64, indirectMax: 3}

func NewBiomes() PalettedContainer[Biome] {
	return PalettedContainer[Biome]{kind: biomePalette}
}

//go:generate go run ./gen biomes

//go:embed biomes.json
var biomes []byte

type BiomeInfo struct {
	ID               Biome      `json:"id"`
	Name             Identifier `json:"name"`
	Temperature      float32    `json:"temperature"`
	Dimension        Identifier `json:"dimension"`
	HasPrecipitation bool       `json:"has_precipitation"`
}

type biomeCatalog struct {
	byID   map[Biome]BiomeInfo
	byName map[Identifier]BiomeInfo
}

var loadBiomes = sync.OnceValue(func() biomeCatalog {
	var entries []BiomeInfo
	if err := json.Unmarshal(biomes, &entries); err != nil {
		panic(fmt.Sprintf("gocraft: embedded biome data is invalid: %v", err))
	}

	catalog := biomeCatalog{
		byID:   make(map[Biome]BiomeInfo, len(entries)),
		byName: make(map[Identifier]BiomeInfo, len(entries)),
	}
	for _, biome := range entries {
		catalog.byID[biome.ID] = biome
		catalog.byName[biome.Name] = biome
	}

	return catalog
})

func BiomeOf(id Biome) (BiomeInfo, bool) {
	biome, ok := loadBiomes().byID[id]

	return biome, ok
}

func BiomeNamed(name Identifier) (BiomeInfo, bool) {
	biome, ok := loadBiomes().byName[name]

	return biome, ok
}
