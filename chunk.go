package gocraft

type ChunkSection struct {
	blockCount Short
	blocks     PalettedContainer[BlockState]
	biomes     PalettedContainer[Biome]
}

func (s *ChunkSection) Decode(r *Reader) error {
	s.blocks = NewBlockStates()
	s.biomes = NewBiomes()

	return DecodeAll(r, &s.blockCount, &s.blocks, &s.biomes)
}

func (s ChunkSection) Empty() bool {
	return s.blockCount == 0
}

func (s ChunkSection) Block(x, y, z int) BlockState {
	return s.blocks.Get(y<<8 | z<<4 | x)
}

func (s ChunkSection) Biome(x, y, z int) Biome {
	return s.biomes.Get(y>>2<<4 | z>>2<<2 | x>>2)
}

type ChunkColumn struct {
	X        int32
	Z        int32
	minY     int
	sections []ChunkSection
}

func NewChunkColumn(x, z int32, minY, height int) *ChunkColumn {
	return &ChunkColumn{
		X:        x,
		Z:        z,
		minY:     minY,
		sections: make([]ChunkSection, height/16),
	}
}

func (c *ChunkColumn) Decode(r *Reader) error {
	for i := range c.sections {
		if err := c.sections[i].Decode(r); err != nil {
			return err
		}
	}

	return nil
}

func (c *ChunkColumn) Section(index int) *ChunkSection {
	return &c.sections[index]
}

func (c *ChunkColumn) Block(x, y, z int) BlockState {
	offset := y - c.minY

	return c.sections[offset>>4].Block(x, offset&15, z)
}

func (c *ChunkColumn) Biome(x, y, z int) Biome {
	offset := y - c.minY

	return c.sections[offset>>4].Biome(x, offset&15, z)
}
