package gocraft

type Biome int32

var biomes = paletteKind{entries: 64, indirectMax: 3}

func NewBiomes() PalettedContainer[Biome] {
	return PalettedContainer[Biome]{kind: biomes}
}
