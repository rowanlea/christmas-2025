package art

import (
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hth/internal/tree"
)

var sizes = []int{3, 4, 16, 48}

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
	bg, bulbs := Load(7)
	if b := bg.Bounds(); b.Dx() != 1200 || b.Dy() != 1200 {
		t.Fatalf("background is %v", b)
	}
	want := color.RGBA{R: 0x0A, G: 0x0E, B: 0x14, A: 0xFF}
	for _, pt := range []image.Point{{0, 0}, {600, 600}, {1199, 1199}} {
		if got := rgb(bg.At(pt.X, pt.Y)); got != want {
			t.Errorf("pixel %v = %v, want %v", pt, got, want)
		}
	}
	if len(bulbs) != 28 {
		t.Fatalf("%d bulbs, want 28", len(bulbs))
	}
	i := 0
	for r := 0; r < 7; r++ {
		for p := 0; p <= r; p++ {
			x, y := tree.Centre(7, r, p)
			if bulbs[i].X != x || bulbs[i].Y != y || bulbs[i].R != tree.Radius(7) {
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
	white := func(im *image.Paletted) int {
		count := 0
		for y := 560; y < 600; y++ {
			for x := 0; x < 600; x++ {
				if rgb(im.At(x, y)) == (color.RGBA{255, 255, 255, 255}) {
					count++
				}
			}
		}
		return count
	}
	if white(g.Image[2]) == 0 {
		t.Errorf("captioned frame has no white caption pixels")
	}
	if white(g.Image[1]) != 0 {
		t.Errorf("uncaptioned frame has white pixels in the caption area")
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
