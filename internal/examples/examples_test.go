package examples

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadExamples(t *testing.T) {
	exs, err := Load("../../examples")
	if err != nil {
		t.Fatal(err)
	}
	if len(exs) != 8 {
		t.Fatalf("got %d examples, want 8", len(exs))
	}
	if exs[0].Name != "1-lights-on" || exs[0].Size != 3 || exs[0].Expected != "RRGRGW" {
		t.Errorf("first example = %+v", exs[0])
	}
}

func TestMissingExpectIsError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.txt"), []byte("SIZE 3\nFILL R\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(dir)
	if err == nil || !strings.Contains(err.Error(), "x.txt") {
		t.Fatalf("err = %v, want error naming x.txt", err)
	}
}

func TestEmptyDirIsError(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(dir)
	if err == nil || !strings.Contains(err.Error(), "no .txt examples") {
		t.Fatalf("err = %v, want no .txt examples error", err)
	}
}

func TestMissingSizeIsError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "y.txt"), []byte("# expect: R\nFILL R\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(dir)
	if err == nil || !strings.Contains(err.Error(), "y.txt") || !strings.Contains(err.Error(), "SIZE") {
		t.Fatalf("err = %v, want SIZE error naming y.txt", err)
	}
}
