package main

import (
	"bytes"
	"strings"
	"testing"

	"hth/internal/program/programtest"
)

func TestHthTestReportsFailures(t *testing.T) {
	fake := programtest.FakePath(t)
	var out, errb bytes.Buffer
	code := run([]string{"test", "../../examples", "--", fake, "echo-k"}, &out, &errb)
	if code != 1 {
		t.Fatalf("exit = %d, want 1; stderr: %s", code, errb.String())
	}
	s := strings.TrimSpace(out.String())
	if !strings.HasSuffix(s, "0/8 passed") {
		t.Errorf("stdout = %q", s)
	}
	if !strings.Contains(s, "FAIL 1-lights-on: wrong bulbs at (0,0)") {
		t.Errorf("missing diff line in %q", s)
	}
}
