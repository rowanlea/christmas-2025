package animate

import (
	"context"
	"image"
	"image/gif"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"hth/internal/program"
	"hth/internal/program/programtest"
	"hth/internal/script"
)

const size = 5

func plan(t *testing.T, steps, limit int) script.Plan {
	t.Helper()
	text := "SIZE 5\n" + strings.Repeat("STEP\n", steps)
	s, err := script.Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	return script.Sample(s, limit)
}

func TestAnimateOrdersFramesAndWritesGIF(t *testing.T) {
	fake := programtest.FakePath(t)
	path := filepath.Join(t.TempDir(), "a.gif")
	final := strings.Repeat("K", 15)
	frames, err := Animate(context.Background(), []string{fake, "count"}, plan(t, 10, 120), size, final, "c", nil, path)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 10 {
		t.Fatalf("got %d frames, want 10", len(frames))
	}
	for i, f := range frames {
		if got := strings.IndexByte(f, 'R'); got != i+1 {
			t.Errorf("frame %d: R at %d, want %d", i, got, i+1)
		}
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
	if len(g.Image) != 11 {
		t.Errorf("gif has %d images, want 11", len(g.Image))
	}
}

func TestAnimateSkipsSlowFrames(t *testing.T) {
	old := program.Timeout
	program.Timeout = 200 * time.Millisecond
	t.Cleanup(func() { program.Timeout = old })
	fake := programtest.FakePath(t)
	p := plan(t, 10, 120)
	want := 0
	for _, f := range p.Frames {
		if strings.Count(f.Script, "\n") <= 5 {
			want++
		}
	}
	start := time.Now()
	frames, err := Animate(context.Background(), []string{fake, "slow-after", "5"}, p, size, strings.Repeat("K", 15), "c", nil, filepath.Join(t.TempDir(), "a.gif"))
	if err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed >= 3*time.Second {
		t.Errorf("took %s, want < 3s", elapsed)
	}
	if want == 0 || len(frames) != want {
		t.Errorf("got %d frames, want %d", len(frames), want)
	}
}

func TestAnimateRunsInParallel(t *testing.T) {
	fake := programtest.FakePath(t)
	start := time.Now()
	frames, err := Animate(context.Background(), []string{fake, "sleep-ms", "300"}, plan(t, 16, 120), size, strings.Repeat("K", 15), "c", nil, filepath.Join(t.TempDir(), "a.gif"))
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 16 {
		t.Fatalf("got %d frames, want 16", len(frames))
	}
	if elapsed := time.Since(start); elapsed >= 1500*time.Millisecond {
		t.Errorf("took %s, want < 1.5s", elapsed)
	}
}

func TestCaptionFormatsThousands(t *testing.T) {
	for n, want := range map[int64]string{7: "7 passes later", 1234: "1,234 passes later", 1000000: "1,000,000 passes later"} {
		if got := Caption(n); got != want {
			t.Errorf("Caption(%d) = %q, want %q", n, got, want)
		}
	}
}

func whitePixels(img *image.Paletted) int {
	n := 0
	for y := 0; y < 600; y++ {
		for x := 0; x < 600; x++ {
			if r, g, b, _ := img.At(x, y).RGBA(); r>>8 == 255 && g>>8 == 255 && b>>8 == 255 {
				n++
			}
		}
	}
	return n
}

func TestAnimateCaptionsTheFinalFrameAfterATrailingSkip(t *testing.T) {
	fake := programtest.FakePath(t)
	final := strings.Repeat("K", 15) // no white bulbs: white pixels are caption ink
	for _, tc := range []struct {
		skipped int64
		inked   bool
	}{{0, false}, {1_234, true}} {
		p := plan(t, 4, 120)
		p.SkippedAtEnd = tc.skipped
		path := filepath.Join(t.TempDir(), "a.gif")
		if _, err := Animate(context.Background(), []string{fake, "count"}, p, size, final, "c", nil, path); err != nil {
			t.Fatal(err)
		}
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		g, err := gif.DecodeAll(f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		last := g.Image[len(g.Image)-1]
		if got := whitePixels(last) > 0; got != tc.inked {
			t.Errorf("skipped %d: final frame captioned = %v, want %v", tc.skipped, got, tc.inked)
		}
	}
}
