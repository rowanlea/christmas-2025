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
	wellScale = 0.8
	capScale  = 0.38
)

func drawColours() []color.RGBA {
	return []color.RGBA{SocketColour, wellColour}
}

// DrawFrame draws the tree for size onto dst: the grid background scaled by
// scale (nearest-neighbour), then bulb i of line on entry i of the grid's bulb
// table in its state colour. Bulb 0 is a five-point star and the rest are
// discs. Grid entries beyond the line, and characters outside RGBWK, are drawn
// as unlit sockets. dst must use Palette(size).
func DrawFrame(dst *image.Paletted, size int, line string, scale float64) {
	grid := Grid(size)
	e := load(grid)
	b := base(e, grid, dst, scale)
	for y := 0; y < b.Rect.Dy(); y++ {
		copy(dst.Pix[y*dst.Stride:y*dst.Stride+b.Rect.Dx()], b.Pix[y*b.Stride:y*b.Stride+b.Rect.Dx()])
	}
	var index [256]uint8
	var known [256]bool
	for _, s := range []byte("RGBWK") {
		index[s] = uint8(dst.Palette.Index(tree.Colour(s)))
		known[s] = true
	}
	socket := uint8(dst.Palette.Index(SocketColour))
	well := uint8(dst.Palette.Index(wellColour))
	for i, bulb := range e.bulbs {
		cx, cy, r := bulb.X*scale, bulb.Y*scale, bulb.R*scale
		if i >= len(line) || !known[line[i]] {
			fillDisc(dst, cx, cy, r*wellScale, well)
			fillDisc(dst, cx, cy, r*capScale, socket)
			continue
		}
		if i == 0 {
			fillStar(dst, cx, cy, r*starScale, index[line[i]])
		} else {
			fillDisc(dst, cx, cy, r, index[line[i]])
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

func fillDisc(dst *image.Paletted, cx, cy, r float64, idx uint8) {
	scanBox(dst, cx-r, cy-r, cx+r, cy+r, func(x, y float64) bool {
		dx, dy := x-cx, y-cy
		return dx*dx+dy*dy <= r*r
	}, idx)
}

const (
	starScale = 1.3
	starInner = 0.5
)

func fillStar(dst *image.Paletted, cx, cy, outer float64, idx uint8) {
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
	}, idx)
}

// scanBox sets every pixel of dst in the box whose centre satisfies inside.
func scanBox(dst *image.Paletted, x0, y0, x1, y1 float64, inside func(x, y float64) bool, idx uint8) {
	b := dst.Rect
	for py := max(int(math.Floor(y0))-1, 0); py <= min(int(math.Ceil(y1))+1, b.Dy()-1); py++ {
		for px := max(int(math.Floor(x0))-1, 0); px <= min(int(math.Ceil(x1))+1, b.Dx()-1); px++ {
			if inside(float64(px)+0.5, float64(py)+0.5) {
				dst.Pix[py*dst.Stride+px] = idx
			}
		}
	}
}
