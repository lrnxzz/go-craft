package gocraft

import "fmt"

type BlockState int32

const Air BlockState = 0

type paletteKind struct {
	entries     int
	indirectMax int
}

var blockStates = paletteKind{entries: 4096, indirectMax: 8}

func (k paletteKind) longs(bitsPerEntry int) int {
	if bitsPerEntry == 0 {
		return 0
	}

	perLong := 64 / bitsPerEntry

	return (k.entries + perLong - 1) / perLong
}

type PalettedContainer[T ~int32] struct {
	kind         paletteKind
	bitsPerEntry int
	palette      Slice[VarInt]
	data         Slice[Long]
}

func NewBlockStates() PalettedContainer[BlockState] {
	return PalettedContainer[BlockState]{kind: blockStates}
}

func (c PalettedContainer[T]) Len() int {
	return c.kind.entries
}

func (c PalettedContainer[T]) Get(index int) T {
	if c.bitsPerEntry == 0 {
		return T(c.palette[0])
	}

	perLong := 64 / c.bitsPerEntry
	long := uint64(c.data[index/perLong])
	value := long >> uint(index%perLong*c.bitsPerEntry) & (uint64(1)<<c.bitsPerEntry - 1)

	if c.palette != nil {
		return T(c.palette[value])
	}

	return T(value)
}

func (c *PalettedContainer[T]) Decode(r *Reader) error {
	var bitsPerEntry UByte
	if err := bitsPerEntry.Decode(r); err != nil {
		return err
	}
	c.bitsPerEntry = int(bitsPerEntry)

	switch {
	case c.bitsPerEntry == 0:
		var value VarInt
		if err := value.Decode(r); err != nil {
			return err
		}
		c.palette = Slice[VarInt]{value}
	case c.bitsPerEntry <= c.kind.indirectMax:
		if err := c.palette.Decode(r); err != nil {
			return err
		}
	}

	if err := c.data.Decode(r); err != nil {
		return err
	}

	expected := c.kind.longs(c.bitsPerEntry)
	if len(c.data) != expected {
		return r.fail(fmt.Errorf("gocraft: paletted container has %d longs, want %d", len(c.data), expected))
	}

	return nil
}
