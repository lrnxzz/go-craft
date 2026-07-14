package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

const (
	tileSize    = 16
	jarVersion  = "1.20.4"
	manifestURL = "https://launchermeta.mojang.com/mc/game/version_manifest_v2.json"
	modelsURL   = "https://raw.githubusercontent.com/PrismarineJS/minecraft-assets/master/data/1.20.2/blocks_models.json"
)

type blockModel struct {
	Parent   string            `json:"parent"`
	Textures map[string]string `json:"textures"`
}

type faceNames struct {
	up, down, side string
}

type faceTiles struct {
	Up   int `json:"up"`
	Down int `json:"down"`
	Side int `json:"side"`
}

type atlasFile struct {
	Tile    int                  `json:"tile"`
	Columns int                  `json:"columns"`
	Rows    int                  `json:"rows"`
	Blocks  map[string]faceTiles `json:"blocks"`
}

func atlasCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "atlas <version>",
		Short: "Generate viewer/assets/{atlas.png,blocks.json} from a codec's Minecraft textures",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			assets := fmt.Sprintf("codec/v%s/assets", args[0])

			jar, err := cached(assets+"/client.jar", fetchClientJar)
			if err != nil {
				return err
			}
			rawModels, err := cached(assets+"/blocks_models.json", func() ([]byte, error) { return fetch(modelsURL) })
			if err != nil {
				return err
			}

			var models map[string]blockModel
			if err := json.Unmarshal(rawModels, &models); err != nil {
				return err
			}
			textures, err := readTextures(jar)
			if err != nil {
				return err
			}

			faces := resolveFaces(models, textures)
			index := assignTiles(faces)

			if err := os.MkdirAll("viewer/assets", 0o755); err != nil {
				return err
			}
			if err := writeAtlas("viewer/assets/atlas.png", textures, index); err != nil {
				return err
			}
			if err := writeBlocks("viewer/assets/blocks.json", faces, index); err != nil {
				return err
			}

			cmd.Printf("atlas: %d blocks, %d tiles\n", len(faces), len(index))

			return nil
		},
	}
}

func cached(pathname string, download func() ([]byte, error)) ([]byte, error) {
	if data, err := os.ReadFile(pathname); err == nil {
		return data, nil
	}

	data, err := download()
	if err != nil {
		return nil, err
	}

	return data, os.WriteFile(pathname, data, 0o644)
}

func readTextures(jar []byte) (map[string]image.Image, error) {
	archive, err := zip.NewReader(bytes.NewReader(jar), int64(len(jar)))
	if err != nil {
		return nil, err
	}

	const prefix = "assets/minecraft/textures/block/"
	textures := map[string]image.Image{}
	for _, file := range archive.File {
		if !strings.HasPrefix(file.Name, prefix) || !strings.HasSuffix(file.Name, ".png") {
			continue
		}

		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		img, err := png.Decode(reader)
		reader.Close()
		if err != nil || img.Bounds().Dx() != tileSize {
			continue
		}

		textures[strings.TrimSuffix(strings.TrimPrefix(file.Name, prefix), ".png")] = img
	}

	return textures, nil
}

func fetchClientJar() ([]byte, error) {
	raw, err := fetch(manifestURL)
	if err != nil {
		return nil, err
	}

	var manifest struct {
		Versions []struct {
			ID  string `json:"id"`
			URL string `json:"url"`
		} `json:"versions"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, err
	}

	var versionURL string
	for _, version := range manifest.Versions {
		if version.ID == jarVersion {
			versionURL = version.URL
		}
	}
	if versionURL == "" {
		return nil, fmt.Errorf("gen: version %s not in manifest", jarVersion)
	}

	raw, err = fetch(versionURL)
	if err != nil {
		return nil, err
	}

	var version struct {
		Downloads struct {
			Client struct {
				URL string `json:"url"`
			} `json:"client"`
		} `json:"downloads"`
	}
	if err := json.Unmarshal(raw, &version); err != nil {
		return nil, err
	}

	return fetch(version.Downloads.Client.URL)
}

func fetch(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gen: GET %s: %s", url, response.Status)
	}

	return io.ReadAll(response.Body)
}

func resolveFaces(models map[string]blockModel, textures map[string]image.Image) map[string]faceNames {
	faces := map[string]faceNames{}
	for name := range models {
		merged := mergeTextures(models, name)

		up := lookup(merged, textures, "up", "top", "end", "all", "side", "particle")
		down := lookup(merged, textures, "down", "bottom", "end", "all", "side", "particle")
		side := lookup(merged, textures, "north", "side", "all", "end", "particle")
		if up == "" || down == "" || side == "" {
			continue
		}

		faces[name] = faceNames{up: up, down: down, side: side}
	}

	return faces
}

func mergeTextures(models map[string]blockModel, name string) map[string]string {
	merged := map[string]string{}
	for name != "" {
		current, ok := models[name]
		if !ok {
			break
		}
		for key, value := range current.Textures {
			if _, exists := merged[key]; !exists {
				merged[key] = value
			}
		}
		name = trimNamespace(current.Parent)
	}

	return merged
}

func lookup(merged map[string]string, textures map[string]image.Image, keys ...string) string {
	for _, key := range keys {
		ref, ok := merged[key]
		if !ok {
			continue
		}
		name := dereference(merged, ref, 0)
		if _, exists := textures[name]; exists {
			return name
		}
	}

	return ""
}

func dereference(merged map[string]string, ref string, depth int) string {
	if depth > 8 {
		return ""
	}
	if strings.HasPrefix(ref, "#") {
		next, ok := merged[ref[1:]]
		if !ok {
			return ""
		}

		return dereference(merged, next, depth+1)
	}

	return trimNamespace(ref)
}

func trimNamespace(ref string) string {
	ref = strings.TrimPrefix(ref, "minecraft:")
	ref = strings.TrimPrefix(ref, "block/")

	return ref
}

func assignTiles(faces map[string]faceNames) map[string]int {
	used := map[string]bool{}
	for _, face := range faces {
		used[face.up], used[face.down], used[face.side] = true, true, true
	}

	names := make([]string, 0, len(used))
	for name := range used {
		names = append(names, name)
	}
	sort.Strings(names)

	index := make(map[string]int, len(names))
	for i, name := range names {
		index[name] = i
	}

	return index
}

func writeAtlas(pathname string, textures map[string]image.Image, index map[string]int) error {
	columns := atlasColumns(len(index))
	rows := (len(index) + columns - 1) / columns
	canvas := image.NewRGBA(image.Rect(0, 0, columns*tileSize, rows*tileSize))

	for name, i := range index {
		originX, originY := (i%columns)*tileSize, (i/columns)*tileSize
		src := textures[name]
		tinted := foliage(name)
		for y := range tileSize {
			for x := range tileSize {
				pixel := src.At(src.Bounds().Min.X+x, src.Bounds().Min.Y+y)
				canvas.Set(originX+x, originY+y, tint(pixel, tinted))
			}
		}
	}

	file, err := os.Create(pathname)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, canvas)
}

func writeBlocks(pathname string, faces map[string]faceNames, index map[string]int) error {
	columns := atlasColumns(len(index))
	data := atlasFile{
		Tile:    tileSize,
		Columns: columns,
		Rows:    (len(index) + columns - 1) / columns,
		Blocks:  make(map[string]faceTiles, len(faces)),
	}
	for name, face := range faces {
		data.Blocks[name] = faceTiles{Up: index[face.up], Down: index[face.down], Side: index[face.side]}
	}

	encoded, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(pathname, encoded, 0o644)
}

func atlasColumns(tiles int) int {
	return int(math.Ceil(math.Sqrt(float64(tiles))))
}

func foliage(name string) bool {
	return strings.Contains(name, "grass_block_top") ||
		strings.Contains(name, "leaves") ||
		strings.Contains(name, "grass") ||
		strings.Contains(name, "fern") ||
		strings.Contains(name, "vine")
}

func tint(pixel color.Color, tinted bool) color.Color {
	if !tinted {
		return pixel
	}

	r, g, b, a := pixel.RGBA()

	return color.RGBA{
		R: uint8(uint32(r>>8) * 124 / 255),
		G: uint8(uint32(g>>8) * 189 / 255),
		B: uint8(uint32(b>>8) * 84 / 255),
		A: uint8(a >> 8),
	}
}
