// Package art holds the precomputed viewer art (tree backgrounds and bulb
// tables) and draws light-tree frames and animated GIFs from them.
package art

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"sort"
	"sync"

	"hth/internal/tree"
)

//go:embed assets/*
var assets embed.FS

// Assets returns the embedded asset files, rooted at the assets folder.
func Assets() fs.FS {
	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}
	return sub
}

// Canvas is the side of the square background and bulb coordinate space.
const Canvas = 1200

// BgColour is the plain background colour (#0A0E14).
var BgColour = color.RGBA{R: 0x0A, G: 0x0E, B: 0x14, A: 0xFF}

// Bulb is one bulb's centre and radius on the 1200 x 1200 canvas.
type Bulb struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	R float64 `json:"r"`
}

type entry struct {
	bg      image.Image
	bulbs   []Bulb
	colours []color.RGBA
}

var (
	mu    sync.Mutex
	cache = map[int]*entry{}
)

// Layout returns the bulb table implied by the render layout, in answer order.
func Layout(size int) []Bulb {
	bulbs := make([]Bulb, 0, tree.BulbCount(size))
	for r := 0; r < size; r++ {
		for p := 0; p <= r; p++ {
			x, y := tree.Centre(size, r, p)
			bulbs = append(bulbs, Bulb{X: x, Y: y, R: tree.Radius(size)})
		}
	}
	return bulbs
}

// Grid returns the size of the bulb grid a tree of size rows is drawn on: 16
// for 1..16 rows and 48 for 17..48. Bulb i of the answer sits on entry i of
// the grid's bulb table, since answer order runs row by row from the star.
// Other sizes have no grid and are drawn on their own render layout.
func Grid(size int) int {
	switch {
	case size >= 1 && size <= 16:
		return 16
	case size >= 17 && size <= 48:
		return 48
	}
	return size
}

// Load returns the 1200 x 1200 background and bulb table of the grid that a
// tree of size rows is drawn on. Sizes without a grid get a plain #0A0E14
// background and the bulbs of their own render layout.
func Load(size int) (bg image.Image, bulbs []Bulb) {
	e := load(Grid(size))
	return e.bg, append([]Bulb(nil), e.bulbs...)
}

func load(size int) *entry {
	mu.Lock()
	defer mu.Unlock()
	if e, ok := cache[size]; ok {
		return e
	}
	e, err := readAssets(size)
	if err != nil {
		panic(fmt.Sprintf("art: embedded assets for size %d: %v", size, err))
	}
	if e == nil {
		plain := image.NewPaletted(image.Rect(0, 0, Canvas, Canvas), color.Palette{BgColour})
		e = &entry{bg: plain, bulbs: Layout(size), colours: []color.RGBA{BgColour}}
	} else {
		e.colours = distinctColours(e.bg)
	}
	cache[size] = e
	return e
}

func readAssets(size int) (*entry, error) {
	pngData, err := assets.ReadFile(fmt.Sprintf("assets/tree-%d.png", size))
	if err != nil {
		return nil, nil
	}
	jsonData, err := assets.ReadFile(fmt.Sprintf("assets/bulbs-%d.json", size))
	if err != nil {
		return nil, fmt.Errorf("missing bulb table: %w", err)
	}
	bg, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, err
	}
	if b := bg.Bounds(); b.Dx() != Canvas || b.Dy() != Canvas {
		return nil, fmt.Errorf("background is %v, want %dx%d", b, Canvas, Canvas)
	}
	var bulbs []Bulb
	if err := json.Unmarshal(jsonData, &bulbs); err != nil {
		return nil, err
	}
	if len(bulbs) != tree.BulbCount(size) {
		return nil, fmt.Errorf("%d bulbs, want %d", len(bulbs), tree.BulbCount(size))
	}
	return &entry{bg: bg, bulbs: bulbs}, nil
}

func distinctColours(img image.Image) []color.RGBA {
	seen := map[color.RGBA]bool{}
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			seen[toRGBA(img.At(x, y))] = true
		}
	}
	out := make([]color.RGBA, 0, len(seen))
	for c := range seen {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.R != b.R {
			return a.R < b.R
		}
		if a.G != b.G {
			return a.G < b.G
		}
		return a.B < b.B
	})
	return out
}

func toRGBA(c color.Color) color.RGBA {
	r, g, b, _ := c.RGBA()
	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 0xFF}
}

// MaxBackgroundColours is the most distinct colours an asset PNG may use. The
// GIF palette adds the caption bar colour, the five bulb colours and the
// colours DrawFrame draws (17 in all), and a GIF palette holds 256.
const MaxBackgroundColours = 239

func backgroundColours(size int) []color.RGBA {
	return load(size).colours
}

// Palette returns the GIF palette for a tree of size rows: the grid
// background's own colours, the caption bar colour, the five bulb colours
// (white doubles as caption text colour) and the colours DrawFrame adds around
// bulbs and sockets, each once.
func Palette(size int) color.Palette {
	var p color.Palette
	seen := map[color.RGBA]bool{}
	add := func(c color.RGBA) {
		if !seen[c] {
			seen[c] = true
			p = append(p, c)
		}
	}
	for _, c := range backgroundColours(Grid(size)) {
		add(c)
	}
	add(BgColour)
	for _, s := range []byte("RGBWK") {
		add(toRGBA(tree.Colour(s)))
	}
	for _, c := range drawColours() {
		add(c)
	}
	return p
}
