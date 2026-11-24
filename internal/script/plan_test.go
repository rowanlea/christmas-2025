package script

import (
	"strings"
	"testing"
	"time"
)

func mustParse(t *testing.T, text string) Script {
	t.Helper()
	s, err := Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestCutDownInsideLoop(t *testing.T) {
	s := mustParse(t, "SIZE 4\nSET 0 0 R\nREPEAT 5\nCHASE\nCYCLE\nEND\n")
	got := cutDown(s.Size, s.Body, []pathEl{{idx: 1, pass: 2}, {idx: 0}})
	want := "SIZE 4\nSET 0 0 R\nREPEAT 2\nCHASE\nCYCLE\nEND\nCHASE\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestCutDownNested(t *testing.T) {
	s := mustParse(t, "SIZE 4\nA\nREPEAT 3\nB\nREPEAT 4\nC\nD\nEND\nE\nEND\n")
	got := cutDown(s.Size, s.Body, []pathEl{{idx: 1, pass: 1}, {idx: 1, pass: 2}, {idx: 1}})
	want := "SIZE 4\nA\nREPEAT 1\nB\nREPEAT 4\nC\nD\nEND\nE\nEND\nB\nREPEAT 2\nC\nD\nEND\nC\nD\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSampleBillionLoop(t *testing.T) {
	s := load(t, "6-silent-night.txt")
	p := Sample(s, 120)
	if len(p.Frames) == 0 || len(p.Frames) > 120 {
		t.Fatalf("%d frames", len(p.Frames))
	}
	if p.SkippedAtEnd != 999_999_900 {
		t.Fatalf("SkippedAtEnd %d", p.SkippedAtEnd)
	}
	for _, f := range p.Frames {
		if f.SkippedBefore != 0 {
			t.Fatalf("unexpected skip %d", f.SkippedBefore)
		}
		fs, err := Parse(f.Script)
		if err != nil {
			t.Fatal(err)
		}
		if fs.Steps().Int64() > 2+3*100 {
			t.Fatalf("frame beyond first 100 passes: %d steps", fs.Steps().Int64())
		}
	}
}

func TestSampleSkipThenMore(t *testing.T) {
	s := mustParse(t, "SIZE 4\nREPEAT 1000000000\nA\nEND\nREPEAT 1000000000\nB\nEND\n")
	p := Sample(s, 120)
	if len(p.Frames) > 120 || p.SkippedAtEnd != 999_999_900 {
		t.Fatalf("%d frames, end skip %d", len(p.Frames), p.SkippedAtEnd)
	}
	var after *Frame
	for i := range p.Frames {
		if p.Frames[i].SkippedBefore != 0 {
			after = &p.Frames[i]
		}
	}
	if after == nil || after.SkippedBefore != 999_999_900 {
		t.Fatalf("no frame after the first skip: %+v", p.Frames)
	}
	if !strings.HasPrefix(after.Script, "SIZE 4\nREPEAT 1000000000\nA\nEND\n") {
		t.Fatalf("script %q", after.Script)
	}
}

func TestSampleShortScript(t *testing.T) {
	s := mustParse(t, "SIZE 3\nA\nB\nC\nD\nE\n")
	p := Sample(s, 120)
	if len(p.Frames) != 5 || p.SkippedAtEnd != 0 {
		t.Fatalf("%+v", p)
	}
	for _, f := range p.Frames {
		if f.SkippedBefore != 0 {
			t.Fatal("skip in short script")
		}
	}
}

func TestSampleSpreadsEvenlyAndKeepsEnds(t *testing.T) {
	s := mustParse(t, "SIZE 3\nREPEAT 100\nREPEAT 5\nA\nEND\nEND\n")
	p := Sample(s, 120)
	if len(p.Frames) != 120 {
		t.Fatalf("%d frames", len(p.Frames))
	}
	last := mustParse(t, p.Frames[119].Script)
	if last.Steps().Int64() != 500 {
		t.Fatalf("last frame is step %d", last.Steps().Int64())
	}
}

func TestSampleCandidateCap(t *testing.T) {
	s := mustParse(t, "SIZE 3\nREPEAT 1000000000\nREPEAT 1000000000\nREPEAT 1000000000\nA\nEND\nEND\nEND\n")
	p := Sample(s, 120)
	if len(p.Frames) > 120 {
		t.Fatalf("%d frames", len(p.Frames))
	}
	if p.SkippedAtEnd != 9_092_999_090_700 {
		t.Fatalf("SkippedAtEnd %d", p.SkippedAtEnd)
	}
	total := p.SkippedAtEnd
	for _, f := range p.Frames {
		total += f.SkippedBefore
	}
	if total != 10_100_998_989_900 {
		t.Fatalf("skips not conserved: %d", total)
	}
}

func TestSampleManySkipsStaysWithinLimit(t *testing.T) {
	text := "SIZE 3\n" + strings.Repeat("REPEAT 1000\nA\nEND\n", 300)
	p := Sample(mustParse(t, text), 120)
	if len(p.Frames) != 120 {
		t.Fatalf("%d frames", len(p.Frames))
	}
	if !strings.HasPrefix(p.Frames[0].Script, "SIZE 3\nA\n") || p.Frames[0].SkippedBefore != 0 {
		t.Fatalf("first frame %q", p.Frames[0].Script)
	}
	last := mustParse(t, p.Frames[119].Script)
	if last.Steps().Int64() != 299*1000+100 {
		t.Fatalf("last frame is step %d", last.Steps().Int64())
	}
	total := p.SkippedAtEnd
	for _, f := range p.Frames {
		total += f.SkippedBefore
	}
	if total != 300*900 {
		t.Fatalf("skips not conserved: %d", total)
	}
}

func TestEmptyLoopsDoNotHang(t *testing.T) {
	for _, body := range []string{
		"REPEAT 1000000000\nEND\n",
		"REPEAT 1000000000\nREPEAT 0\nX\nEND\nEND\n",
		"REPEAT 1000000000\nREPEAT 1000000000\nREPEAT 1000000000\nEND\nEND\nEND\n",
	} {
		s := mustParse(t, "SIZE 3\n"+body+"A\n")
		done := make(chan [2]Plan, 1)
		go func() { done <- [2]Plan{Every(s), Sample(s, 120)} }()
		select {
		case ps := <-done:
			for _, p := range ps {
				if len(p.Frames) != 1 || p.SkippedAtEnd != 0 || p.Frames[0].SkippedBefore != 0 {
					t.Fatalf("%q: %+v", body, p)
				}
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("%q: planner hangs", body)
		}
	}
}

func TestEveryRefusesOver500Steps(t *testing.T) {
	p := Every(mustParse(t, "SIZE 3\nREPEAT 501\nA\nEND\n"))
	if len(p.Frames) != 0 {
		t.Fatalf("%d frames", len(p.Frames))
	}
	if len(Every(mustParse(t, "SIZE 3\nREPEAT 500\nA\nEND\n")).Frames) != 500 {
		t.Fatal("500 steps should be kept")
	}
}

func TestEveryKeepsAllSteps(t *testing.T) {
	text := "SIZE 4\n" + strings.Repeat("S\n", 20) + "REPEAT 10\nA\nB\nEND\n"
	p := Every(mustParse(t, text))
	if len(p.Frames) != 40 {
		t.Fatalf("%d frames", len(p.Frames))
	}
}

func TestCutDownRoundTrips(t *testing.T) {
	text := "SIZE 4\nX\nREPEAT 3\nA\nREPEAT 2\nB\nC\nEND\nD\nEND\nY\nREPEAT 0\nZ\nEND\nREPEAT 2\nREPEAT 2\nQ\nEND\nEND\n"
	p := Every(mustParse(t, text))
	for i, f := range p.Frames {
		fs, err := Parse(f.Script)
		if err != nil {
			t.Fatalf("frame %d: %v", i, err)
		}
		if got := fs.Steps().Int64(); got != int64(i+1) {
			t.Fatalf("frame %d: Steps %d\n%s", i, got, f.Script)
		}
	}
	if len(p.Frames) != 24 {
		t.Fatalf("%d frames", len(p.Frames))
	}
}
