package gocraft

type Biome int32

var biomePalette = paletteKind{entries: 64, indirectMax: 3}

func NewBiomes() PalettedContainer[Biome] {
	return PalettedContainer[Biome]{kind: biomePalette}
}

type BiomeInfo struct {
	ID               Biome      `json:"id"`
	Name             Identifier `json:"name"`
	Temperature      float32    `json:"temperature"`
	Dimension        Identifier `json:"dimension"`
	HasPrecipitation bool       `json:"has_precipitation"`
}
