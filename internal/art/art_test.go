package art

import (
	"image"
	"image/color"
	"image/gif"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hth/internal/tree"
)

var sizes = []int{16, 48}

func rgb(c color.Color) color.RGBA {
	r, g, b, _ := c.RGBA()
	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 255}
}

func TestAssetsPresent(t *testing.T) {
	for _, n := range sizes {
		bg, bulbs := Load(n)
		if b := bg.Bounds(); b.Dx() != 1200 || b.Dy() != 1200 {
			t.Errorf("size %d: background is %v, want 1200x1200", n, b)
		}
		if len(bulbs) != tree.BulbCount(n) {
			t.Fatalf("size %d: %d bulbs, want %d", n, len(bulbs), tree.BulbCount(n))
		}
		i := 0
		for r := 0; r < n; r++ {
			for p := 0; p <= r; p++ {
				if p > 0 && bulbs[i].X <= bulbs[i-1].X {
					t.Errorf("size %d bulb %d: x does not increase within a row", n, i)
				}
				if i > 0 && bulbs[i].Y < bulbs[i-1].Y {
					t.Errorf("size %d bulb %d: y decreases", n, i)
				}
				i++
			}
		}
	}
}

func TestBulbTablesMatchRenderLayout(t *testing.T) {
	for _, n := range sizes {
		_, bulbs := Load(n)
		i := 0
		for r := 0; r < n; r++ {
			for p := 0; p <= r; p++ {
				x, y := tree.Centre(n, r, p)
				if abs(bulbs[i].X-x) > 0.01 || abs(bulbs[i].Y-y) > 0.01 || abs(bulbs[i].R-tree.Radius(n)) > 0.01 {
					t.Fatalf("size %d bulb %d: %+v, want (%.2f, %.2f, %.2f)", n, i, bulbs[i], x, y, tree.Radius(n))
				}
				i++
			}
		}
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func TestBackgroundAvoidsPalette(t *testing.T) {
	for _, n := range sizes {
		bg, _ := Load(n)
		seen := map[color.RGBA]bool{}
		b := bg.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				seen[rgb(bg.At(x, y))] = true
			}
		}
		for c := range seen {
			for _, s := range []byte("RGBWK") {
				p := tree.Colour(s)
				dr, dg, db := int(c.R)-int(p.R), int(c.G)-int(p.G), int(c.B)-int(p.B)
				if d := dr*dr + dg*dg + db*db; d < 3000 {
					t.Fatalf("size %d: background colour %v is within %d of palette %c", n, c, d, s)
				}
			}
		}
	}
}

func TestBackgroundColourBudget(t *testing.T) {
	for _, n := range sizes {
		if got := len(backgroundColours(n)); got > 240 {
			t.Errorf("size %d: %d background colours, want <= 240", n, got)
		}
	}
}

func TestFallbackSize(t *testing.T) {
	const n = 49
	bg, bulbs := Load(n)
	if b := bg.Bounds(); b.Dx() != 1200 || b.Dy() != 1200 {
		t.Fatalf("background is %v", b)
	}
	want := color.RGBA{R: 0x0A, G: 0x0E, B: 0x14, A: 0xFF}
	for _, pt := range []image.Point{{0, 0}, {600, 600}, {1199, 1199}} {
		if got := rgb(bg.At(pt.X, pt.Y)); got != want {
			t.Errorf("pixel %v = %v, want %v", pt, got, want)
		}
	}
	if len(bulbs) != tree.BulbCount(n) {
		t.Fatalf("%d bulbs, want %d", len(bulbs), tree.BulbCount(n))
	}
	i := 0
	for r := 0; r < n; r++ {
		for p := 0; p <= r; p++ {
			x, y := tree.Centre(n, r, p)
			if bulbs[i].X != x || bulbs[i].Y != y || bulbs[i].R != tree.Radius(n) {
				t.Fatalf("bulb %d = %+v, want centre (%v, %v)", i, bulbs[i], x, y)
			}
			i++
		}
	}
}

func TestDrawFrameColoursBulbs(t *testing.T) {
	const n = 16
	line := strings.Repeat("RGBWK", 40)[:tree.BulbCount(n)]
	dst := image.NewPaletted(image.Rect(0, 0, 600, 600), Palette(n))
	DrawFrame(dst, n, line, 0.5)
	_, bulbs := Load(n)
	for i, b := range bulbs {
		want := tree.Colour(line[i])
		got := rgb(dst.At(int(b.X/2), int(b.Y/2)))
		if got != (color.RGBA{R: want.R, G: want.G, B: want.B, A: 255}) {
			t.Fatalf("bulb %d: pixel %v, want %v", i, got, want)
		}
	}
}

func TestDrawFrameStarAndDiscShapes(t *testing.T) {
	const n = 16
	line := strings.Repeat("R", tree.BulbCount(n))
	dst := image.NewPaletted(image.Rect(0, 0, 600, 600), Palette(n))
	DrawFrame(dst, n, line, 0.5)
	_, bulbs := Load(n)
	red := uint8(dst.Palette.Index(tree.Colour('R')))
	b := bulbs[1]
	for _, d := range [][2]float64{{0.9, 0}, {-0.9, 0}, {0, 0.9}, {0, -0.9}} {
		if dst.ColorIndexAt(int((b.X+d[0]*b.R)/2), int((b.Y+d[1]*b.R)/2)) != red {
			t.Errorf("disc bulb missing pixel at offset %v", d)
		}
	}
	s := bulbs[0]
	// Straight below the star centre lies in the notch between its two lower tips.
	if dst.ColorIndexAt(int(s.X/2), int((s.Y+0.9*s.R)/2)) == red {
		t.Errorf("bulb 0 looks like a disc, want a star")
	}
	if dst.ColorIndexAt(int(s.X/2), int((s.Y-1.2*s.R)/2)) != red {
		t.Errorf("star top tip missing")
	}
}

func TestWriteGIF(t *testing.T) {
	const n = 3
	l := tree.BulbCount(n)
	frames := []string{strings.Repeat("K", l), strings.Repeat("R", l), "RGBWKR"}
	path := filepath.Join(t.TempDir(), "x.gif")
	if err := WriteGIF(path, n, frames, map[int]string{2: "12, 3 passes later"}, "RGBWKR"); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	g, err := gif.DecodeAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Image) != 4 {
		t.Fatalf("%d images, want 4", len(g.Image))
	}
	for i, im := range g.Image {
		if b := im.Bounds(); b.Dx() != 600 || b.Dy() != 600 {
			t.Errorf("image %d is %v", i, b)
		}
	}
	want := []int{10, 10, 10, 200}
	for i, d := range g.Delay {
		if d != want[i] {
			t.Errorf("delays %v, want %v", g.Delay, want)
			break
		}
	}
}

func TestWriteGIFRejectsUnknownCaptionGlyph(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.gif")
	err := WriteGIF(path, 3, []string{"KKKKKK"}, map[int]string{0: "Z!"}, "KKKKKK")
	if err == nil {
		t.Fatal("want error for caption glyph outside the font")
	}
}

func TestFontHoldsOnlyTheCaptionAlphabet(t *testing.T) {
	const want = "0123456789, paselrt"
	for _, c := range want {
		if _, ok := glyphs[c]; !ok {
			t.Errorf("no glyph for %q", c)
		}
	}
	if len(glyphs) != len(want) {
		t.Errorf("font has %d glyphs, want %d", len(glyphs), len(want))
	}
}

func decodeGIF(t *testing.T, size int, frames []string, captions map[int]string, final string) *gif.GIF {
	t.Helper()
	path := filepath.Join(t.TempDir(), "x.gif")
	if err := WriteGIF(path, size, frames, captions, final); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	g, err := gif.DecodeAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestWriteGIFCaptionBar(t *testing.T) {
	const n = 3
	line := strings.Repeat("K", tree.BulbCount(n)) // no white bulbs: white pixels can only be caption ink
	g := decodeGIF(t, n, []string{line, line}, map[int]string{1: "12, 3 passes later"}, line)
	plain, captioned := g.Image[0], g.Image[1]

	const barTop = 600 - (7*fontScale + 2*captionPad)
	white := color.RGBA{255, 255, 255, 255}
	minX, maxX, minY, maxY, ink := 600, -1, 600, -1, 0
	differs := false
	for y := 0; y < 600; y++ {
		for x := 0; x < 600; x++ {
			if rgb(plain.At(x, y)) == white {
				t.Fatalf("uncaptioned frame has white pixel at (%d,%d)", x, y)
			}
			if rgb(captioned.At(x, y)) == white {
				ink++
				minX, maxX, minY, maxY = min(minX, x), max(maxX, x), min(minY, y), max(maxY, y)
			}
			if y >= barTop && rgb(captioned.At(x, y)) != rgb(plain.At(x, y)) {
				differs = true
			}
			if y < barTop && rgb(captioned.At(x, y)) != rgb(plain.At(x, y)) {
				t.Fatalf("caption changed pixel (%d,%d) above the bar", x, y)
			}
		}
	}
	if !differs || ink == 0 {
		t.Fatalf("caption not drawn: bar differs=%v, ink pixels=%d", differs, ink)
	}
	if minY < barTop+captionPad || maxY >= barTop+captionPad+7*fontScale {
		t.Errorf("ink rows %d..%d outside text rows of the bar", minY, maxY)
	}
	if centre := (minX + maxX) / 2; centre < 297 || centre > 303 {
		t.Errorf("caption ink spans x %d..%d, centre %d, want about 300", minX, maxX, centre)
	}
}

func TestGridSize(t *testing.T) {
	for size, want := range map[int]int{1: 16, 3: 16, 4: 16, 16: 16, 17: 48, 30: 48, 48: 48, 0: 0, 49: 49} {
		if got := Grid(size); got != want {
			t.Errorf("Grid(%d) = %d, want %d", size, got, want)
		}
	}
}

func TestOnlyGridAssetsRemain(t *testing.T) {
	entries, err := fs.ReadDir(Assets(), ".")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	want := ".gitattributes README.md bulbs-16.json bulbs-48.json tree-16.png tree-48.png"
	if got := strings.Join(names, " "); got != want {
		t.Errorf("assets = %s, want %s", got, want)
	}
}

func TestSmallTreesLoadTheGrid(t *testing.T) {
	for size, grid := range map[int]int{3: 16, 4: 16, 9: 16, 17: 48, 30: 48} {
		gbg, gbulbs := Load(grid)
		bg, bulbs := Load(size)
		if bg != gbg || len(bulbs) != len(gbulbs) {
			t.Fatalf("Load(%d): %d bulbs, want the %d grid's %d", size, len(bulbs), grid, len(gbulbs))
		}
		for i := range bulbs {
			if bulbs[i] != gbulbs[i] {
				t.Fatalf("Load(%d) bulb %d = %+v, want %+v", size, i, bulbs[i], gbulbs[i])
			}
		}
	}
}

func TestDrawFramePutsASmallAnswerOnTheGrid(t *testing.T) {
	const n = 4
	line := "RGBWKRGBWK"
	dst := image.NewPaletted(image.Rect(0, 0, 600, 600), Palette(n))
	DrawFrame(dst, n, line, 0.5)
	_, grid := Load(16)
	if len(grid) != 136 {
		t.Fatalf("%d grid bulbs, want 136", len(grid))
	}
	for i, b := range grid {
		got := rgb(dst.At(int(b.X/2), int(b.Y/2)))
		if i < len(line) {
			want := tree.Colour(line[i])
			if got != (color.RGBA{R: want.R, G: want.G, B: want.B, A: 255}) {
				t.Errorf("grid entry %d: pixel %v, want bulb %c", i, got, line[i])
			}
		} else if got != SocketColour {
			t.Errorf("grid entry %d: pixel %v, want socket %v", i, got, SocketColour)
		}
	}
}

func TestSocketNeverLooksLikeABulbState(t *testing.T) {
	for _, s := range []byte("RGBWK") {
		p := tree.Colour(s)
		dr, dg, db := int(SocketColour.R)-int(p.R), int(SocketColour.G)-int(p.G), int(SocketColour.B)-int(p.B)
		if d := dr*dr + dg*dg + db*db; d < 3000 {
			t.Errorf("socket colour %v is within %d of %c", SocketColour, d, s)
		}
	}
}
