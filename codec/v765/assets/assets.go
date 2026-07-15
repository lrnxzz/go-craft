package assets

import _ "embed"

//go:embed blocks.json
var Blocks []byte

//go:embed biomes.json
var Biomes []byte

//go:embed items.json
var Items []byte
