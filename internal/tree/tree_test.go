package tree

import (
	"math"
	"testing"
)

func TestBulbCount(t *testing.T) {
	for size, want := range map[int]int{3: 6, 16: 136, 48: 1176} {
		if got := BulbCount(size); got != want {
			t.Errorf("BulbCount(%d) = %d, want %d", size, got, want)
		}
	}
}

func TestValidate(t *testing.T) {
	if err := Validate("RRGRGW", 3); err != nil {
		t.Errorf("valid line rejected: %v", err)
	}
	err := Validate("RRGRG", 3)
	if err == nil || err.Error() != "answer has 5 characters, expected 6 for SIZE 3" {
		t.Errorf("short line: got %v", err)
	}
	err = Validate("RRGRGx", 3)
	if err == nil || err.Error() != "answer has invalid character 'x' at position 5 (allowed: RGBWK)" {
		t.Errorf("bad char: got %v", err)
	}
}

func TestCentre(t *testing.T) {
	x, y := Centre(16, 0, 0)
	if x != 600 || y != 120 {
		t.Errorf("Centre(16,0,0) = (%v,%v)", x, y)
	}
	x, y = Centre(16, 15, 0)
	wantY := 120 + 15*0.866*1000/15
	if math.Abs(x-100) > 1e-9 || math.Abs(y-wantY) > 1e-9 {
		t.Errorf("Centre(16,15,0) = (%v,%v), want (100,%v)", x, y, wantY)
	}
}

func TestDiff(t *testing.T) {
	got := Diff("RRGRGW", "RRGRGK", 3)
	if len(got) != 1 || got[0] != [2]int{2, 2} {
		t.Errorf("Diff = %v", got)
	}
	if got := Diff("RRGRGW", "RRGRGW", 3); len(got) != 0 {
		t.Errorf("identical Diff = %v", got)
	}
}

func TestWiringSnakesFromTheStar(t *testing.T) {
	want := []int{0, 1, 2, 5, 4, 3, 6, 7, 8, 9}
	got := Wiring(4)
	if len(got) != len(want) {
		t.Fatalf("Wiring(4) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Wiring(4) = %v, want %v", got, want)
		}
	}
	if n := len(Wiring(48)); n != BulbCount(48) {
		t.Errorf("Wiring(48) has %d entries, want %d", n, BulbCount(48))
	}
}
