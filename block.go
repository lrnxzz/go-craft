package gocraft

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"
)

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

//go:generate go run ./gen

//go:embed blocks.json
var blocks []byte

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

type Block struct {
	Name         Identifier `json:"name"`
	MinState     BlockState `json:"minStateId"`
	MaxState     BlockState `json:"maxStateId"`
	DefaultState BlockState `json:"defaultState"`
	Properties   []Property `json:"states"`
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

var loadBlocks = sync.OnceValue(func() []Block {
	var catalog []Block
	if err := json.Unmarshal(blocks, &catalog); err != nil {
		panic(fmt.Sprintf("gocraft: embedded block data is invalid: %v", err))
	}

	return catalog
})

var blocksByName = sync.OnceValue(func() map[Identifier]Block {
	catalog := loadBlocks()

	byName := make(map[Identifier]Block, len(catalog))
	for _, block := range catalog {
		byName[block.Name] = block
	}

	return byName
})

func BlockOf(state BlockState) (Block, bool) {
	catalog := loadBlocks()

	index := sort.Search(len(catalog), func(i int) bool {
		return catalog[i].MaxState >= state
	})
	if index < len(catalog) && catalog[index].MinState <= state {
		return catalog[index], true
	}

	return Block{}, false
}

func BlockNamed(name Identifier) (Block, bool) {
	block, ok := blocksByName()[name]

	return block, ok
}
