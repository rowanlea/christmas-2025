// Command genart writes the first-draft viewer art assets (tree backgrounds and
// bulb tables) that the organisers then finish by hand.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"hth/internal/art"
)

var sizes = []int{3, 4, 16, 48}

// Decoration colours stay at least squared RGB distance 3000 from every bulb
// colour (the art tests enforce it), so lit bulbs never blend into the tree.
var (
	bgColour = art.BgColour
	tiers    = []color.RGBA{
		{R: 0x04, G: 0x24, B: 0x0F, A: 0xFF},
		{R: 0x0A, G: 0x5A, B: 0x28, A: 0xFF},
		{R: 0x0F, G: 0x7A, B: 0x3A, A: 0xFF},
	}
	trunkColour = color.RGBA{R: 0x6B, G: 0x44, B: 0x23, A: 0xFF}
	standColour = color.RGBA{R: 0x6B, G: 0x72, B: 0x80, A: 0xFF}
)

func main() {
	out := flag.String("out", filepath.Join("internal", "art", "assets"), "folder to write the assets into")
	force := flag.Bool("force", false, "overwrite existing assets, including hand-finished ones")
	flag.Parse()
	if err := run(*out, *force); err != nil {
		fmt.Fprintln(os.Stderr, "genart:", err)
		os.Exit(1)
	}
}

func run(dir string, force bool) error {
	if !force {
		var exist []string
		for _, n := range sizes {
			for _, f := range fileNames(n) {
				if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
					exist = append(exist, f)
				}
			}
		}
		if len(exist) > 0 {
			return fmt.Errorf("refusing to overwrite existing assets in %s (%s); they may be hand-finished, pass -force to replace them", dir, strings.Join(exist, ", "))
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, n := range sizes {
		pngName, jsonName := fileNames(n)[0], fileNames(n)[1]
		img, err := encodeTree(n)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, pngName), img, 0o644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, jsonName), bulbTable(n), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func fileNames(n int) []string {
	return []string{fmt.Sprintf("tree-%d.png", n), fmt.Sprintf("bulbs-%d.json", n)}
}

func bulbTable(size int) []byte {
	var b bytes.Buffer
	b.WriteString("[\n")
	bulbs := art.Layout(size)
	for i, bu := range bulbs {
		sep := ","
		if i == len(bulbs)-1 {
			sep = ""
		}
		fmt.Fprintf(&b, "  {\"x\":%s,\"y\":%s,\"r\":%s}%s\n", num(bu.X), num(bu.Y), num(bu.R), sep)
	}
	b.WriteString("]\n")
	return b.Bytes()
}

func num(f float64) string {
	s := fmt.Sprintf("%.2f", f)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}

func encodeTree(size int) ([]byte, error) {
	pal := color.Palette{bgColour}
	for _, c := range tiers {
		pal = append(pal, c)
	}
	pal = append(pal, trunkColour, standColour)
	img := image.NewPaletted(image.Rect(0, 0, art.Canvas, art.Canvas), pal)
	drawTree(img, size)
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const (
	trunkTop, trunkBottom = 1040, 1130
	standTop, standBottom = 1130, 1165
)

func drawTree(img *image.Paletted, size int) {
	idx := func(c color.RGBA) uint8 { return uint8(img.Palette.Index(c)) }
	bulbs := art.Layout(size)
	rad := bulbs[0].R
	rows := make([][]art.Bulb, size)
	i := 0
	for r := 0; r < size; r++ {
		rows[r] = bulbs[i : i+r+1]
		i += r + 1
	}

	fillRect(img, 600-40, trunkTop, 600+40, trunkBottom, idx(trunkColour))
	fillTrapezoid(img, 600-110, 600+110, 600-150, 600+150, standTop, standBottom, idx(standColour))

	nt := 3
	if size >= 16 {
		nt = 4
	}
	// Bottom tier first, so each higher tier overlaps the one below it.
	for t := nt - 1; t >= 0; t-- {
		first, last := t*size/nt, (t+1)*size/nt-1
		if last < first {
			continue
		}
		yb := math.Min(rows[last][0].Y+rad*1.6, trunkTop+20)
		half := 0.0
		for r := first; r <= last; r++ {
			half = math.Max(half, rowHalfWidth(rows[r], rad))
		}
		wb := half + rad*0.6 + 14
		// Apex high enough that the triangle encloses every bulb of the tier.
		ya := rows[first][0].Y
		for r := first; r <= last; r++ {
			need := rowHalfWidth(rows[r], rad)
			if wb > need {
				ya = math.Min(ya, (wb*rows[r][0].Y-need*yb)/(wb-need))
			}
		}
		fillTriangle(img, 600, ya, wb, yb, idx(tiers[(nt-1-t)%len(tiers)]))
	}
}

func rowHalfWidth(row []art.Bulb, rad float64) float64 {
	return math.Max(math.Abs(row[0].X-600), math.Abs(row[len(row)-1].X-600)) + rad*1.3
}

func fillRect(img *image.Paletted, x0, y0, x1, y1 int, idx uint8) {
	for y := max(y0, 0); y < min(y1, art.Canvas); y++ {
		for x := max(x0, 0); x < min(x1, art.Canvas); x++ {
			img.Pix[y*img.Stride+x] = idx
		}
	}
}

func fillTrapezoid(img *image.Paletted, tl, tr, bl, br float64, y0, y1 int, idx uint8) {
	for y := y0; y < y1; y++ {
		f := float64(y-y0) / float64(y1-y0)
		fillRect(img, int(tl+(bl-tl)*f), y, int(tr+(br-tr)*f), y+1, idx)
	}
}

// fillTriangle fills the isosceles triangle with apex (cx, ya) and base at yb
// with half-width wb, clipped to the canvas.
func fillTriangle(img *image.Paletted, cx, ya, wb, yb float64, idx uint8) {
	for y := max(int(math.Ceil(ya)), 0); y < min(int(yb), art.Canvas); y++ {
		w := wb * (float64(y) + 0.5 - ya) / (yb - ya)
		fillRect(img, int(math.Round(cx-w)), y, int(math.Round(cx+w)), y+1, idx)
	}
}
