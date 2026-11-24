package tree

import (
	"bytes"
	"flag"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the PNG fixtures in testdata")

func TestRenderPixels(t *testing.T) {
	line := "RGBWKR"
	data, err := Render(line, 3)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 1200 || b.Dy() != 1200 {
		t.Fatalf("size = %v", b)
	}
	if got := color.NRGBAModel.Convert(img.At(0, 0)); got != (color.NRGBA{0, 0, 0, 255}) {
		t.Errorf("background = %v", got)
	}
	palette := map[byte]color.NRGBA{
		'R': {0xFF, 0, 0, 255}, 'G': {0, 0xFF, 0, 255}, 'B': {0, 0, 0xFF, 255},
		'W': {0xFF, 0xFF, 0xFF, 255}, 'K': {0x30, 0x30, 0x30, 255},
	}
	i := 0
	for r := 0; r < 3; r++ {
		for p := 0; p <= r; p++ {
			x, y := Centre(3, r, p)
			got := color.NRGBAModel.Convert(img.At(int(x), int(y)))
			if got != palette[line[i]] {
				t.Errorf("bulb %d at (%d,%d) = %v, want %v", i, int(x), int(y), got, palette[line[i]])
			}
			i++
		}
	}
}

func cycle(n int) string {
	return strings.Repeat("RGBWK", n/5+1)[:n]
}

func TestFixturesUpToDate(t *testing.T) {
	fixtures := []struct {
		name string
		line string
		size int
	}{
		{"tree-3-mixed", "RGBWKR", 3},
		{"tree-16-cycle", cycle(136), 16},
		{"tree-48-cycle", cycle(1176), 48},
	}
	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			got, err := Render(f.line, f.size)
			if err != nil {
				t.Fatal(err)
			}
			checkFixture(t, f.name, f.line, got)
		})
	}
	t.Run("blank-16", func(t *testing.T) {
		img := image.NewNRGBA(image.Rect(0, 0, 1200, 1200))
		for i := 3; i < len(img.Pix); i += 4 {
			img.Pix[i] = 255
		}
		var buf bytes.Buffer
		enc := png.Encoder{CompressionLevel: png.BestCompression}
		if err := enc.Encode(&buf, img); err != nil {
			t.Fatal(err)
		}
		checkFixture(t, "blank-16", strings.Repeat(".", 136), buf.Bytes())
	})
}

func checkFixture(t *testing.T, name, line string, png []byte) {
	t.Helper()
	pngPath := filepath.Join("testdata", name+".png")
	txtPath := filepath.Join("testdata", name+".txt")
	if *update {
		if err := os.WriteFile(pngPath, png, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(txtPath, []byte(line), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, png) {
		t.Errorf("%s is stale; run go test ./internal/tree -update", pngPath)
	}
	txt, err := os.ReadFile(txtPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(txt) != line {
		t.Errorf("%s does not match the expected line", txtPath)
	}
}
