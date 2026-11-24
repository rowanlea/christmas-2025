// Command genart writes the viewer art assets: the tree backgrounds and bulb
// tables of the 16 and 48 bulb grids.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hth/internal/art"
)

var sizes = []int{16, 48}

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
