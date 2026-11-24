package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesAllAssets(t *testing.T) {
	dir := t.TempDir()
	if err := run(dir, false); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"3", "4", "16", "48"} {
		for _, f := range []string{"tree-" + n + ".png", "bulbs-" + n + ".json"} {
			if st, err := os.Stat(filepath.Join(dir, f)); err != nil || st.Size() == 0 {
				t.Errorf("%s missing or empty: %v", f, err)
			}
		}
	}
}

func TestRunRefusesToOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	hand := filepath.Join(dir, "tree-16.png")
	if err := os.WriteFile(hand, []byte("hand edited"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := run(dir, false)
	if err == nil || !strings.Contains(err.Error(), "-force") {
		t.Fatalf("err = %v, want refusal mentioning -force", err)
	}
	if b, _ := os.ReadFile(hand); string(b) != "hand edited" {
		t.Fatalf("existing asset was modified")
	}
	if _, err := os.Stat(filepath.Join(dir, "tree-3.png")); err == nil {
		t.Fatalf("refusal must not write any asset")
	}
	if err := run(dir, true); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(hand); string(b) == "hand edited" {
		t.Fatalf("-force did not overwrite")
	}
}
