package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"

	"hth/internal/art"
	"hth/internal/tree"
)

// Every colour here stays at least squared RGB distance 3000 from the bulb
// colours (the art tests enforce it), so a lit bulb never blends into the
// decoration. Each ramp runs dark to light.
var (
	night  = ramp(0x05070B, 0x07090E, 0x0A0E14)
	flakes = ramp(0x46546E, 0x8A9AB8)
	snow   = ramp(0x5C6C88, 0x8C9CB4, 0xB4C2D4, 0xD2DCE8)
	greens = ramp(0x02140A, 0x03240F, 0x00401A, 0x0A5A28, 0x137034, 0x2A8A40, 0x4CA850, 0x8CC860, 0xB8D8B0)
	bark   = ramp(0x1E0E02, 0x5A2E08, 0x7A4214, 0x9A5C24)
	brass  = ramp(0x1A1000, 0x6A4A12, 0x9C7420, 0xC8A040, 0xE8CC70)
	wire   = rgb(0x0A0C08)
	hollow = rgb(0x050607)
)

func rgb(v uint32) color.RGBA {
	return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xFF}
}

func ramp(vs ...uint32) []color.RGBA {
	out := make([]color.RGBA, len(vs))
	for i, v := range vs {
		out[i] = rgb(v)
	}
	return out
}

// The tree is painted in 2 x 2 pixel blocks, so the GIF's half-size
// nearest-neighbour copy keeps every block and its dither pattern.
const block = 2

var bayer = [4][4]float64{{0, 8, 2, 10}, {12, 4, 14, 6}, {3, 11, 1, 9}, {15, 7, 13, 5}}

// shade picks the colour for v in [0, 1] along r, ordered-dithering between
// neighbouring steps by the block position (bx, by).
func shade(r []color.RGBA, v float64, bx, by int) color.RGBA {
	v = math.Max(0, math.Min(1, v)) * float64(len(r)-1)
	i := int(v)
	if v-float64(i) > (bayer[by&3][bx&3]+0.5)/16 {
		i++
	}
	return r[min(i, len(r)-1)]
}

// hash is a fixed pseudo-random value in [0, 1) for an integer pair.
func hash(a, b int) float64 {
	h := uint32(a)*0x9E3779B1 ^ uint32(b)*0x85EBCA77
	h ^= h >> 15
	h *= 0x2C1B3C6D
	h ^= h >> 12
	h *= 0x297A2D39
	h ^= h >> 15
	return float64(h) / (1 << 32)
}

func frac(f float64) float64 { return f - math.Floor(f) }

const (
	cx        = 600.0
	apexY     = 100.0
	baseY     = 1040.0
	tierCount = 7
	slope     = 0.5774 // tan 30 degrees: the bulb triangle's sides
	trunkTop  = 960.0
	potRim    = 1078.0
	potTop    = 1094.0
	potBottom = 1158.0
)

// tier is one layer of drooping boughs, from its hidden top t down to its
// scalloped bottom edge around b.
type tier struct {
	k          int
	t, b       float64
	flare      float64
	droop      float64
	depth      float64
	period     float64
	bottomHalf float64
}

func tiers() []tier {
	var out []tier
	prev := apexY
	for k := 0; k < tierCount; k++ {
		b := apexY + (baseY-apexY)*math.Pow(float64(k+1)/tierCount, 0.92)
		t := apexY
		if k > 0 {
			t = prev - 0.6*(b-prev)
		}
		tr := tier{k: k, t: t, b: b, flare: 26 + 0.045*(b-apexY)}
		tr.droop = 10 + 0.025*(b-apexY)
		tr.depth = 9 + 0.12*(b-t)
		tr.bottomHalf = envelope(b) + 6 + tr.flare
		tr.period = 2 * tr.bottomHalf / float64(3+2*k)
		out = append(out, tr)
		prev = b
	}
	return out
}

// envelope is the half-width that holds every bulb, its socket ring and a
// margin at height y.
func envelope(y float64) float64 {
	return math.Max(0, y-apexY)*slope + 22
}

func (tr tier) half(y float64) float64 {
	s := math.Max(0, (y-tr.t)/(tr.b-tr.t))
	teeth := 16 * math.Pow(frac(y/24+float64(tr.k)*0.37), 2)
	return envelope(y) + 6 + tr.flare*math.Pow(s, 1.5) + teeth
}

func (tr tier) bough(dx float64) float64 {
	u := dx / tr.period
	return u - math.Round(u)
}

func (tr tier) bottom(dx float64) float64 {
	n := dx / tr.bottomHalf
	f := tr.bough(dx)
	return tr.b + tr.droop*n*n - tr.depth*4*f*f + 3*hash(int(dx/4), tr.k)
}

func (tr tier) contains(x, y float64) bool {
	dx := x - cx
	return y >= tr.t && math.Abs(dx) <= min(tr.half(y), tr.bottomHalf) && y <= tr.bottom(dx)
}

// light is the tier's brightness at (x, y): darker under the tier above,
// lighter towards the bough tips and on the left, with ridges fanning out
// along each bough and needle streaks along them.
func (tr tier) light(x, y float64, above *tier, bx, by int) float64 {
	dx := x - cx
	s := (y - tr.t) / (tr.b - tr.t)
	n := dx / tr.half(y)
	v := 0.02 + 0.52*s - 0.2*n - 0.12*n*n
	ray := dx * (tr.b - tr.t) / math.Max(y-tr.t, 8)
	f := tr.bough(ray)
	v += 0.22*(1-4*f*f)*s - 0.1 - 0.25*math.Pow(4*f*f, 3)*s
	v += 0.14 * (hash(int(math.Floor(ray/5)), tr.k+7) - 0.5)
	v += 0.1 * (hash(bx, by) - 0.5)
	if above != nil {
		if d := y - above.bottom(dx); d >= 0 && d < 48 {
			v -= 0.32 * (1 - d/48)
		}
	}
	if e := tr.bottom(dx) - y; e < 8 {
		v += 0.24 * (1 - e/8)
	}
	return v
}

func groundY(x float64) float64 {
	dx := x - cx
	return 1122 - 14*math.Cos(dx/300) + 5*math.Sin(x/83)
}

func potHalf(y float64) float64 {
	if y < potTop {
		return 114
	}
	return 102 - 18*(y-potTop)/(potBottom-potTop)
}

// scenery is the colour of the background, snow, trunk, pot or tree at the
// centre (x, y) of block (bx, by).
func scenery(ts []tier, x, y float64, bx, by int) color.RGBA {
	dx := x - cx
	for i, tr := range ts {
		if tr.contains(x, y) {
			var above *tier
			if i > 0 {
				above = &ts[i-1]
			}
			return shade(greens, tr.light(x, y, above, bx, by), bx, by)
		}
	}
	if y >= potRim && y < potBottom && math.Abs(dx) <= potHalf(y) {
		n := dx / potHalf(y)
		v := 0.85 - 0.75*math.Abs(n+0.35)
		if y < potTop {
			v += 0.12
		} else if y > 1116 && y < 1124 {
			v -= 0.3
		}
		if y > potBottom-8 {
			v -= 0.25
		}
		return shade(brass, v+0.08*(hash(bx, by)-0.5), bx, by)
	}
	if y >= trunkTop && y < potRim+4 && math.Abs(dx) <= 34+(y-trunkTop)*0.05 {
		n := dx / 34
		v := 0.62 - 0.45*n + 0.35*(hash(int(math.Floor((dx+40)/6)), 3)-0.5)
		v -= 0.5 * math.Max(0, 1-(y-trunkTop)/90)
		return shade(bark, v, bx, by)
	}
	if g := groundY(x); y >= g {
		v := 0.92 - (y-g)/180
		if e := (dx/190)*(dx/190) + ((y-potBottom)/20)*((y-potBottom)/20); e < 1 {
			v -= 0.5 * (1 - e)
		}
		if y-g < 4 {
			v += 0.1
		}
		return shade(snow, v+0.06*(hash(bx, by)-0.5), bx, by)
	}
	if h := hash(bx*3+1, by*5+2); h < 0.004 {
		if h < 0.0008 {
			return flakes[1]
		}
		return flakes[0]
	}
	d := math.Hypot(dx, y-560) / 860
	return shade(night, 1.15-d*1.1, bx, by)
}

func palette() color.Palette {
	var p color.Palette
	for _, r := range [][]color.RGBA{night, flakes, snow, greens, bark, brass} {
		for _, c := range r {
			p = append(p, c)
		}
	}
	return append(p, wire, hollow)
}

func encodeTree(size int) ([]byte, error) {
	img := image.NewPaletted(image.Rect(0, 0, art.Canvas, art.Canvas), palette())
	paint(img, size)
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func paint(img *image.Paletted, size int) {
	idx := func(c color.RGBA) uint8 { return uint8(img.Palette.Index(c)) }
	ts := tiers()
	for by := 0; by < art.Canvas/block; by++ {
		for bx := 0; bx < art.Canvas/block; bx++ {
			c := idx(scenery(ts, float64(bx*block)+1, float64(by*block)+1, bx, by))
			for y := by * block; y < (by+1)*block; y++ {
				for x := bx * block; x < (bx+1)*block; x++ {
					img.Pix[y*img.Stride+x] = c
				}
			}
		}
	}
	bulbs := art.Layout(size)
	drawWire(img, size, bulbs, idx(wire))
	drawSockets(img, bulbs, idx)
}

// drawWire runs a sagging dark wire between the bulbs in wiring order.
func drawWire(img *image.Paletted, size int, bulbs []art.Bulb, c uint8) {
	order := tree.Wiring(size)
	w := math.Max(1.1, bulbs[0].R*0.095)
	for i := 1; i < len(order); i++ {
		a, b := bulbs[order[i-1]], bulbs[order[i]]
		dist := math.Hypot(b.X-a.X, b.Y-a.Y)
		sag := 0.13 * dist
		for step := 0.0; step <= dist; step += 0.5 {
			t := step / dist
			x := a.X + (b.X-a.X)*t
			y := a.Y + (b.Y-a.Y)*t + sag*4*t*(1-t)
			if math.Hypot(x-a.X, y-a.Y) < a.R*socketRing || math.Hypot(x-b.X, y-b.Y) < b.R*socketRing {
				continue
			}
			disc(img, x, y, w, func(int, int) uint8 { return c })
		}
	}
}

// Sockets are small, so a tree with few bulbs still shows its branches.
const (
	socketHollow = 0.42
	socketRing   = 0.62
)

// drawSockets paints a small socket, a dark hollow in a brass ring lit from
// the top left, at every bulb position but the star's.
func drawSockets(img *image.Paletted, bulbs []art.Bulb, idx func(color.RGBA) uint8) {
	h := idx(hollow)
	for _, b := range bulbs[1:] {
		inner := b.R * socketHollow
		outer := math.Max(b.R*socketRing, inner+2)
		disc(img, b.X, b.Y, outer, func(x, y int) uint8 {
			dx, dy := float64(x)+0.5-b.X, float64(y)+0.5-b.Y
			d := math.Hypot(dx, dy)
			if d < inner {
				return h
			}
			lit := -(dx + dy) / (d * math.Sqrt2)
			return idx(shade(brass[1:], 0.5+0.5*lit, x, y))
		})
	}
}

func disc(img *image.Paletted, cx, cy, r float64, colour func(x, y int) uint8) {
	for y := max(int(cy-r)-1, 0); y <= min(int(cy+r)+1, art.Canvas-1); y++ {
		for x := max(int(cx-r)-1, 0); x <= min(int(cx+r)+1, art.Canvas-1); x++ {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if dx*dx+dy*dy <= r*r {
				img.Pix[y*img.Stride+x] = colour(x, y)
			}
		}
	}
}
