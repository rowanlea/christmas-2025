// Package tree validates answer lines and renders them as PNG light trees.
package tree

import "fmt"

const (
	canvas  = 1200
	originX = 600.0
	originY = 120.0
	spanX   = 1000.0
	rowRise = 0.866
	radiusK = 0.32
	allowed = "RGBWK"
)

// BulbCount returns the number of bulbs in a tree of the given number of rows.
func BulbCount(size int) int {
	return size * (size + 1) / 2
}

// Validate checks that line holds exactly one RGBWK character per bulb.
func Validate(line string, size int) error {
	if want := BulbCount(size); len(line) != want {
		return fmt.Errorf("answer has %d characters, expected %d for SIZE %d", len(line), want, size)
	}
	for i := 0; i < len(line); i++ {
		if !isColour(line[i]) {
			return fmt.Errorf("answer has invalid character '%c' at position %d (allowed: %s)", line[i], i, allowed)
		}
	}
	return nil
}

func isColour(c byte) bool {
	for i := 0; i < len(allowed); i++ {
		if allowed[i] == c {
			return true
		}
	}
	return false
}

func spacing(size int) float64 {
	if size < 2 {
		return spanX
	}
	return spanX / float64(size-1)
}

// Centre returns the canvas position of bulb p in row r.
func Centre(size, r, p int) (x, y float64) {
	h := spacing(size)
	return originX + (float64(p)-float64(r)/2)*h, originY + float64(r)*rowRise*h
}

// Radius returns the bulb radius in pixels.
func Radius(size int) float64 {
	return radiusK * spacing(size)
}

// Diff returns the (row, position) of every bulb where the two lines differ,
// in answer order. Only the common prefix of the lines is compared.
func Diff(expected, actual string, size int) [][2]int {
	n := min(len(expected), len(actual), BulbCount(size))
	var out [][2]int
	i := 0
	for r := 0; r < size && i < n; r++ {
		for p := 0; p <= r && i < n; p++ {
			if expected[i] != actual[i] {
				out = append(out, [2]int{r, p})
			}
			i++
		}
	}
	return out
}
