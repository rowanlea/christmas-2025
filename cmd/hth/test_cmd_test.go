package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"hth/internal/program/programtest"
)

func TestHthTestReportsFailures(t *testing.T) {
	fake := programtest.FakePath(t)
	examples, err := filepath.Abs("../../examples")
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	var out, errb bytes.Buffer
	code := run([]string{"test", examples, "--", fake, "echo-k"}, &out, &errb)
	if code != 1 {
		t.Fatalf("exit = %d, want 1; stderr: %s", code, errb.String())
	}
	s := strings.TrimSpace(out.String())
	if !strings.Contains(s, "0/1 passed") {
		t.Errorf("stdout = %q", s)
	}
	if !strings.Contains(s, "FAIL worked-example: wrong bulbs at (0,0)") {
		t.Errorf("missing diff line in %q", s)
	}
}
