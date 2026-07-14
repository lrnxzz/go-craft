package gocraft

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type BlockState int32

type paletteKind struct {
	entries     int
	indirectMax int
	directBits  int
}

var blockStates = paletteKind{entries: 4096, indirectMax: 8, directBits: 15}

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
		if len(c.palette) == 0 {
			return 0
		}

		return T(c.palette[0])
	}

	symbol := c.symbolAt(index)
	if c.palette != nil {
		return T(c.palette[symbol])
	}

	return T(symbol)
}

func (c PalettedContainer[T]) symbolAt(index int) uint64 {
	perLong := 64 / c.bitsPerEntry
	long := uint64(c.data[index/perLong])

	return long >> uint(index%perLong*c.bitsPerEntry) & (uint64(1)<<c.bitsPerEntry - 1)
}

func (c *PalettedContainer[T]) put(index int, symbol uint64) {
	perLong := 64 / c.bitsPerEntry
	shift := uint(index % perLong * c.bitsPerEntry)
	mask := uint64(1)<<c.bitsPerEntry - 1

	long := &c.data[index/perLong]
	*long = Long(uint64(*long)&^(mask<<shift) | (symbol&mask)<<shift)
}

func (c PalettedContainer[T]) paletteSlot(value T) int {
	for slot, entry := range c.palette {
		if T(entry) == value {
			return slot
		}
	}

	return -1
}

func (c *PalettedContainer[T]) Set(index int, value T) {
	if c.bitsPerEntry == 0 {
		if c.Get(0) == value {
			return
		}
		if len(c.palette) == 0 {
			c.palette = Slice[VarInt]{0}
		}
		c.repack(1)
	}

	if c.palette == nil {
		c.put(index, uint64(value))

		return
	}

	slot := c.paletteSlot(value)
	if slot < 0 {
		if len(c.palette) == 1<<c.bitsPerEntry {
			if c.bitsPerEntry < c.kind.indirectMax {
				c.repack(c.bitsPerEntry + 1)
			} else {
				c.toDirect()
				c.put(index, uint64(value))

				return
			}
		}

		slot = len(c.palette)
		c.palette = append(c.palette, VarInt(value))
	}

	c.put(index, uint64(slot))
}

func (c *PalettedContainer[T]) repack(bitsPerEntry int) {
	symbols := make([]uint64, c.kind.entries)
	for index := range symbols {
		if c.bitsPerEntry != 0 {
			symbols[index] = c.symbolAt(index)
		}
	}

	c.bitsPerEntry = bitsPerEntry
	c.data = make(Slice[Long], c.kind.longs(bitsPerEntry))
	for index, symbol := range symbols {
		c.put(index, symbol)
	}
}

func (c *PalettedContainer[T]) toDirect() {
	values := make([]uint64, c.kind.entries)
	for index := range values {
		values[index] = uint64(c.Get(index))
	}

	c.palette = nil
	c.bitsPerEntry = c.kind.directBits
	c.data = make(Slice[Long], c.kind.longs(c.kind.directBits))
	for index, value := range values {
		c.put(index, value)
	}
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

type Property struct {
	Name   string
	Values []string
}

func (p *Property) UnmarshalJSON(data []byte) error {
	var raw propertyData
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	p.Name = raw.Name
	switch raw.Type {
	case "bool":
		p.Values = []string{"true", "false"}
	case "enum":
		p.Values = raw.Values
	default:
		p.Values = make([]string, raw.NumValues)
		for i := range p.Values {
			p.Values[i] = strconv.Itoa(i)
		}
	}

	return nil
}

type propertyData struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	NumValues int      `json:"num_values"`
	Values    []string `json:"values"`
}

type BoundingBox string

const (
	BoundingBoxEmpty BoundingBox = "empty"
	BoundingBoxBlock BoundingBox = "block"
)

type Block struct {
	Name         Identifier  `json:"name"`
	MinState     BlockState  `json:"minStateId"`
	MaxState     BlockState  `json:"maxStateId"`
	DefaultState BlockState  `json:"defaultState"`
	BoundingBox  BoundingBox `json:"boundingBox"`
	Properties   []Property  `json:"states"`
}

func (b Block) Solid() bool {
	return b.BoundingBox == BoundingBoxBlock
}

func (b Block) At(state BlockState) map[string]string {
	values := make(map[string]string, len(b.Properties))

	offset := int(state - b.MinState)
	for i := len(b.Properties) - 1; i >= 0; i-- {
		property := b.Properties[i]
		values[property.Name] = property.Values[offset%len(property.Values)]
		offset /= len(property.Values)
	}

	return values
}
