package tree

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
)

var palette = map[byte]color.NRGBA{
	'R': {R: 0xFF, A: 0xFF},
	'G': {G: 0xFF, A: 0xFF},
	'B': {B: 0xFF, A: 0xFF},
	'W': {R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
	'K': {R: 0x30, G: 0x30, B: 0x30, A: 0xFF},
}

// Render draws the line as a 1200x1200 PNG with hard-edged bulbs, so the
// server can sample bulb centres exactly.
func Render(line string, size int) ([]byte, error) {
	if err := Validate(line, size); err != nil {
		return nil, err
	}
	img := image.NewNRGBA(image.Rect(0, 0, canvas, canvas))
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 0xFF
	}
	rad := Radius(size)
	i := 0
	for r := 0; r < size; r++ {
		for p := 0; p <= r; p++ {
			cx, cy := Centre(size, r, p)
			fill(img, cx, cy, rad, Colour(line[i]))
			i++
		}
	}
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode png: %w", err)
	}
	return buf.Bytes(), nil
}

func fill(img *image.NRGBA, cx, cy, rad float64, c color.NRGBA) {
	x0, x1 := int(math.Floor(cx-rad))-1, int(math.Ceil(cx+rad))+1
	y0, y1 := int(math.Floor(cy-rad))-1, int(math.Ceil(cy+rad))+1
	for y := max(y0, 0); y <= min(y1, canvas-1); y++ {
		for x := max(x0, 0); x <= min(x1, canvas-1); x++ {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if dx*dx+dy*dy <= rad*rad {
				img.SetNRGBA(x, y, c)
			}
		}
	}
}

// Colour returns the bulb colour for an answer character (R, G, B, W or K).
func Colour(state byte) color.NRGBA {
	return palette[state]
}
