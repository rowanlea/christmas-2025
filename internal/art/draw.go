package art

import (
	"image"
	"image/color"
	"math"
	"sync"

	"hth/internal/tree"
)

type baseKey struct {
	size, w, h int
	scale      float64
	colours    int
}

var (
	baseMu    sync.Mutex
	baseCache = map[baseKey]*image.Paletted{}
)

// SocketColour is the cap of an unlit socket: a grid position that holds no
// bulb of the answer. It is far from every bulb colour, Off included.
var SocketColour = color.RGBA{R: 0x6A, G: 0x4A, B: 0x12, A: 0xFF}

// wellColour is the dark hollow around a socket's cap.
var wellColour = color.RGBA{R: 0x05, G: 0x06, B: 0x07, A: 0xFF}

const (
	wellScale = 0.42
	capScale  = 0.26
)

// look is how a bulb state is drawn in a GIF frame: a lit bulb has a halo,
// solid near the bulb and dithered further out, and every bulb has a small
// glint on its glass, so Off reads as an unlit glass bulb.
type look struct {
	lit         bool
	halo, glint color.RGBA
}

var looks = map[byte]look{
	'R': {true, color.RGBA{R: 0xB0, G: 0x20, B: 0x1C, A: 0xFF}, color.RGBA{R: 0xFF, G: 0xA8, B: 0x98, A: 0xFF}},
	'G': {true, color.RGBA{R: 0x2A, G: 0xB0, B: 0x40, A: 0xFF}, color.RGBA{R: 0xC8, G: 0xFF, B: 0xC8, A: 0xFF}},
	'B': {true, color.RGBA{R: 0x28, G: 0x38, B: 0xC8, A: 0xFF}, color.RGBA{R: 0xA8, G: 0xB8, B: 0xFF, A: 0xFF}},
	'W': {true, color.RGBA{R: 0xC8, G: 0xC4, B: 0xA0, A: 0xFF}, color.RGBA{R: 0xFF, G: 0xF0, B: 0xC0, A: 0xFF}},
	'K': {false, color.RGBA{}, color.RGBA{R: 0x74, G: 0x74, B: 0x6C, A: 0xFF}},
}

const (
	haloSolid  = 1.28
	haloOuter  = 1.62
	glintShift = 0.36
	glintScale = 0.22
)

func drawColours() []color.RGBA {
	out := []color.RGBA{SocketColour, wellColour}
	for _, s := range []byte("RGBWK") {
		l := looks[s]
		if l.lit {
			out = append(out, l.halo)
		}
		out = append(out, l.glint)
	}
	return out
}

// DrawFrame draws the tree for size onto dst: the grid background scaled by
// scale (nearest-neighbour), then bulb i of line on entry i of the grid's bulb
// table in its state colour, with a halo around lit bulbs. Bulb 0 is a
// five-point star and the rest are discs. Grid entries beyond the line, and
// characters outside RGBWK, are drawn as unlit sockets. dst must use
// Palette(size).
func DrawFrame(dst *image.Paletted, size int, line string, scale float64) {
	grid := Grid(size)
	e := load(grid)
	b := base(e, grid, dst, scale)
	for y := 0; y < b.Rect.Dy(); y++ {
		copy(dst.Pix[y*dst.Stride:y*dst.Stride+b.Rect.Dx()], b.Pix[y*b.Stride:y*b.Stride+b.Rect.Dx()])
	}
	ix := func(c color.Color) uint8 { return uint8(dst.Palette.Index(c)) }
	state := func(i int) byte {
		if i < len(line) {
			if _, ok := looks[line[i]]; ok {
				return line[i]
			}
		}
		return 0
	}
	shape := func(i int, cx, cy, r, grow float64, idx uint8, checker bool) {
		if i == 0 {
			fillStar(dst, cx, cy, r*starScale*grow, idx, checker)
		} else {
			fillDisc(dst, cx, cy, r*grow, idx, checker)
		}
	}
	for i, bulb := range e.bulbs {
		if l := looks[state(i)]; l.lit {
			cx, cy, r := bulb.X*scale, bulb.Y*scale, bulb.R*scale
			shape(i, cx, cy, r, haloOuter, ix(l.halo), true)
			shape(i, cx, cy, r, haloSolid, ix(l.halo), false)
		}
	}
	for i, bulb := range e.bulbs {
		cx, cy, r := bulb.X*scale, bulb.Y*scale, bulb.R*scale
		s := state(i)
		if s == 0 {
			fillDisc(dst, cx, cy, r*wellScale, ix(wellColour), false)
			fillDisc(dst, cx, cy, r*capScale, ix(SocketColour), false)
			continue
		}
		shape(i, cx, cy, r, 1, ix(tree.Colour(s)), false)
		if i > 0 {
			fillDisc(dst, cx-glintShift*r, cy-glintShift*r, max(glintScale*r, 0.71), ix(looks[s].glint), false)
		}
	}
}

func base(e *entry, size int, dst *image.Paletted, scale float64) *image.Paletted {
	key := baseKey{size, dst.Rect.Dx(), dst.Rect.Dy(), scale, len(dst.Palette)}
	baseMu.Lock()
	defer baseMu.Unlock()
	if b, ok := baseCache[key]; ok {
		return b
	}
	lookup := make(map[color.RGBA]uint8, len(dst.Palette))
	for i, c := range dst.Palette {
		lookup[toRGBA(c)] = uint8(i)
	}
	b := image.NewPaletted(image.Rect(0, 0, key.w, key.h), dst.Palette)
	sb := e.bg.Bounds()
	for y := 0; y < key.h; y++ {
		sy := min(int(float64(y)/scale), sb.Dy()-1)
		for x := 0; x < key.w; x++ {
			sx := min(int(float64(x)/scale), sb.Dx()-1)
			c := toRGBA(e.bg.At(sb.Min.X+sx, sb.Min.Y+sy))
			i, ok := lookup[c]
			if !ok {
				i = uint8(dst.Palette.Index(c))
			}
			b.Pix[y*b.Stride+x] = i
		}
	}
	baseCache[key] = b
	return b
}

func fillDisc(dst *image.Paletted, cx, cy, r float64, idx uint8, checker bool) {
	scanBox(dst, cx-r, cy-r, cx+r, cy+r, func(x, y float64) bool {
		dx, dy := x-cx, y-cy
		return dx*dx+dy*dy <= r*r
	}, idx, checker)
}

const (
	starScale = 1.3
	starInner = 0.5
)

func fillStar(dst *image.Paletted, cx, cy, outer float64, idx uint8, checker bool) {
	var vx, vy [10]float64
	for k := 0; k < 10; k++ {
		rad := outer
		if k%2 == 1 {
			rad = outer * starInner
		}
		a := -math.Pi/2 + float64(k)*math.Pi/5
		vx[k], vy[k] = cx+rad*math.Cos(a), cy+rad*math.Sin(a)
	}
	scanBox(dst, cx-outer, cy-outer, cx+outer, cy+outer, func(x, y float64) bool {
		in := false
		for i, j := 0, 9; i < 10; j, i = i, i+1 {
			if (vy[i] > y) != (vy[j] > y) && x < (vx[j]-vx[i])*(y-vy[i])/(vy[j]-vy[i])+vx[i] {
				in = !in
			}
		}
		return in
	}, idx, checker)
}

// scanBox sets every pixel of dst in the box whose centre satisfies inside,
// or with checker only every other pixel, in a checkerboard.
func scanBox(dst *image.Paletted, x0, y0, x1, y1 float64, inside func(x, y float64) bool, idx uint8, checker bool) {
	b := dst.Rect
	for py := max(int(math.Floor(y0))-1, 0); py <= min(int(math.Ceil(y1))+1, b.Dy()-1); py++ {
		for px := max(int(math.Floor(x0))-1, 0); px <= min(int(math.Ceil(x1))+1, b.Dx()-1); px++ {
			if checker && (px+py)%2 == 1 {
				continue
			}
			if inside(float64(px)+0.5, float64(py)+0.5) {
				dst.Pix[py*dst.Stride+px] = idx
			}
		}
	}
}
