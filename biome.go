package gocraft

type BiomeID int32

var biomePalette = paletteType{
	entries:     64,
	indirectMax: 3,
	directBits:  6,
}

func Biomes() PalettedContainer[BiomeID] {
	return PalettedContainer[BiomeID]{
		paletteType: biomePalette,
	}
}

type Biome struct {
	ID               BiomeID    `json:"id"`
	Name             Identifier `json:"name"`
	Temperature      float32    `json:"temperature"`
	Dimension        Identifier `json:"dimension"`
	HasPrecipitation bool       `json:"has_precipitation"`
}
