package viewer

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"image/png"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765/blocks"
	"github.com/lrnxzz/go-craft/viewer/gpu"
	"github.com/lrnxzz/go-craft/viewer/mesh"
)

//go:embed assets/atlas.png
var atlasImage []byte

//go:embed assets/blocks.json
var atlasMapping []byte

type faceTiles struct {
	Up   int `json:"up"`
	Down int `json:"down"`
	Side int `json:"side"`
}

type atlasFile struct {
	Columns int                              `json:"columns"`
	Rows    int                              `json:"rows"`
	Blocks  map[gocraft.Identifier]faceTiles `json:"blocks"`
}

type Tileset struct {
	atlas  *gpu.Atlas
	blocks map[gocraft.Identifier]faceTiles
}

func LoadTileset() (*Tileset, error) {
	img, err := png.Decode(bytes.NewReader(atlasImage))
	if err != nil {
		return nil, err
	}

	var file atlasFile
	if err := json.Unmarshal(atlasMapping, &file); err != nil {
		return nil, err
	}

	return &Tileset{
		atlas:  gpu.NewAtlas(gpu.NewTexture(img), file.Columns, file.Rows),
		blocks: file.Blocks,
	}, nil
}

func (t *Tileset) Atlas() *gpu.Atlas {
	return t.atlas
}

func (t *Tileset) Tile(state gocraft.BlockState, face mesh.Face) gpu.UV {
	return t.atlas.Tile(t.index(state, face))
}

func (t *Tileset) index(state gocraft.BlockState, face mesh.Face) int {
	block, ok := blocks.Of(state)
	if !ok {
		return 0
	}

	tiles, known := t.blocks[block.Name]
	if !known {
		return 0
	}

	switch face {
	case mesh.Up:
		return tiles.Up
	case mesh.Down:
		return tiles.Down
	default:
		return tiles.Side
	}
}
