package art

import (
	"fmt"
	"image"
)

const (
	glyphW, glyphH = 5, 7
	fontScale      = 3
	captionPad     = 5
)

// glyphs is a 5x7 bitmap font holding only what captions need: digits, comma,
// space and the letters of "passes later". A 1 is ink.
var glyphs = map[rune][glyphH]string{
	'0': {"01110", "10001", "10011", "10101", "11001", "10001", "01110"},
	'1': {"00100", "01100", "00100", "00100", "00100", "00100", "01110"},
	'2': {"01110", "10001", "00001", "00010", "00100", "01000", "11111"},
	'3': {"11110", "00001", "00001", "01110", "00001", "00001", "11110"},
	'4': {"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
	'5': {"11111", "10000", "11110", "00001", "00001", "10001", "01110"},
	'6': {"00110", "01000", "10000", "11110", "10001", "10001", "01110"},
	'7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
	'8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
	'9': {"01110", "10001", "10001", "01111", "00001", "00010", "01100"},
	',': {"00000", "00000", "00000", "00000", "00110", "00100", "01000"},
	' ': {"00000", "00000", "00000", "00000", "00000", "00000", "00000"},
	'p': {"00000", "10110", "11001", "10001", "11110", "10000", "10000"},
	'a': {"00000", "00000", "01110", "00001", "01111", "10001", "01111"},
	's': {"00000", "00000", "01111", "10000", "01110", "00001", "11110"},
	'e': {"00000", "00000", "01110", "10001", "11111", "10000", "01110"},
	'l': {"01100", "00100", "00100", "00100", "00100", "00100", "01110"},
	't': {"00100", "00100", "11111", "00100", "00100", "00101", "00010"},
	'r': {"00000", "00000", "10110", "11001", "10000", "10000", "10000"},
}

// drawCaption fills a bar along the bottom of dst with bg and writes text in
// ink, centred. It fails if text uses a glyph the font lacks.
func drawCaption(dst *image.Paletted, text string, bg, ink uint8) error {
	runes := []rune(text)
	for _, r := range runes {
		if _, ok := glyphs[r]; !ok {
			return fmt.Errorf("caption %q: no glyph for %q", text, r)
		}
	}
	b := dst.Rect
	barH := glyphH*fontScale + 2*captionPad
	for y := b.Dy() - barH; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			dst.Pix[y*dst.Stride+x] = bg
		}
	}
	textW := len(runes)*(glyphW+1)*fontScale - fontScale
	x0, y0 := (b.Dx()-textW)/2, b.Dy()-barH+captionPad
	for i, r := range runes {
		g := glyphs[r]
		for row := 0; row < glyphH; row++ {
			for col := 0; col < glyphW; col++ {
				if g[row][col] != '1' {
					continue
				}
				for dy := 0; dy < fontScale; dy++ {
					for dx := 0; dx < fontScale; dx++ {
						x := x0 + i*(glyphW+1)*fontScale + col*fontScale + dx
						if x >= 0 && x < b.Dx() {
							dst.Pix[(y0+row*fontScale+dy)*dst.Stride+x] = ink
						}
					}
				}
			}
		}
	}
	return nil
}
