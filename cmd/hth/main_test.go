package main

import (
	"bytes"
	"strings"
	"testing"
)

func runCapture(args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestNoArgsPrintsUsage(t *testing.T) {
	code, _, stderr := runCapture()
	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	for _, want := range []string{"hth login <team>", "hth run <puzzle> -- <command…>", "hth test <dir> -- <command…>", "hth show <file> -- <command…>"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("usage missing %q in %q", want, stderr)
		}
	}
}

func TestUnknownCommand(t *testing.T) {
	if code, _, _ := runCapture("frobnicate"); code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
}

func TestRunWithoutSeparator(t *testing.T) {
	code, _, stderr := runCapture("run", "3.2", "python")
	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "--") {
		t.Errorf("stderr should mention --: %q", stderr)
	}
}

func TestVersion(t *testing.T) {
	code, stdout, _ := runCapture("--version")
	if code != 0 || !strings.Contains(stdout, "dev") {
		t.Fatalf("code=%d stdout=%q", code, stdout)
	}
}
