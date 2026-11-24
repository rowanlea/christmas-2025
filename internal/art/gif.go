package art

import (
	"fmt"
	"image"
	"image/gif"
	"os"

	"hth/internal/tree"
)

const (
	gifSide      = 600
	frameDelay   = 10  // 100 ms, in hundredths of a second
	finalHold    = 200 // 2 s
	gifBackScale = float64(gifSide) / Canvas
)

// WriteGIF writes an animated 600 x 600 GIF of the tree for size: one frame
// per entry of frames at 100 ms, then final held for 2 s. captions maps a
// zero-based index into frames to a caption bar for that frame (digits,
// commas, spaces and the letters of "passes later" only).
func WriteGIF(path string, size int, frames []string, captions map[int]string, final string) error {
	pal := Palette(size)
	bg := uint8(pal.Index(BgColour))
	ink := uint8(pal.Index(toRGBA(tree.Colour('W'))))
	out := &gif.GIF{}
	add := func(line, caption string, delay int) error {
		img := image.NewPaletted(image.Rect(0, 0, gifSide, gifSide), pal)
		DrawFrame(img, size, line, gifBackScale)
		if caption != "" {
			if err := drawCaption(img, caption, bg, ink); err != nil {
				return err
			}
		}
		out.Image = append(out.Image, img)
		out.Delay = append(out.Delay, delay)
		return nil
	}
	for i, line := range frames {
		if err := add(line, captions[i], frameDelay); err != nil {
			return err
		}
	}
	if err := add(final, "", finalHold); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := gif.EncodeAll(f, out); err != nil {
		f.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	return f.Close()
}
