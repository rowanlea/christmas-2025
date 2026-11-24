package stubs

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const stubScript = "# starter stub check\nSIZE 16\n"

func TestStubs(t *testing.T) {
	python := ""
	for _, name := range []string{"python3", "python"} {
		if _, err := exec.LookPath(name); err == nil {
			python = name
			break
		}
	}

	tests := []struct {
		name    string
		runtime string
		args    []string
		timeout time.Duration
	}{
		{"python", python, []string{"stubs/python/emulator.py"}, 30 * time.Second},
		{"javascript", "node", []string{"stubs/javascript/emulator.js"}, 30 * time.Second},
		{"java", "java", []string{"stubs/java/Emulator.java"}, 60 * time.Second},
		{"csharp", "dotnet", []string{"run", "--project", "stubs/csharp"}, 300 * time.Second},
		{"go", "go", []string{"run", "./stubs/go"}, 120 * time.Second},
	}

	want := strings.Repeat("K", 136) + "\n"
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.runtime == "" {
				t.Skip("runtime not found on PATH")
			}
			if _, err := exec.LookPath(tc.runtime); err != nil {
				t.Skipf("%s not found on PATH", tc.runtime)
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			cmd := exec.CommandContext(ctx, tc.runtime, tc.args...)
			cmd.Dir = ".."
			cmd.Stdin = strings.NewReader(stubScript)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("run failed: %v\nstderr: %s", err, stderr.String())
			}
			got := strings.ReplaceAll(stdout.String(), "\r\n", "\n")
			if got != want {
				t.Fatalf("stdout = %q, want 136 K and newline", got)
			}
		})
	}
}
