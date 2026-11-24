package program

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"hth/internal/program/programtest"
)

func run(t *testing.T, script string, stderr *bytes.Buffer, args ...string) (string, error) {
	t.Helper()
	if stderr == nil {
		stderr = &bytes.Buffer{}
	}
	return Run(context.Background(), append([]string{programtest.FakePath(t)}, args...), script, stderr)
}

func wantErr(t *testing.T, err error, sub string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), sub) {
		t.Fatalf("error = %v, want containing %q", err, sub)
	}
}

func TestAnswerFromStdout(t *testing.T) {
	got, err := run(t, "SIZE 3\n", nil, "echo-k")
	if err != nil || got != "KKKKKK" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestCRLFAndTrailingSpaceTrimmed(t *testing.T) {
	got, err := run(t, "SIZE 3\n", nil, "crlf")
	if err != nil || got != "KKK" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestDebugOnStdoutRejected(t *testing.T) {
	_, err := run(t, "SIZE 3\n", nil, "debug-stdout")
	wantErr(t, err, "expected 1 answer line, got 2 (print debugging to stderr)")
}

func TestTimeout(t *testing.T) {
	old := Timeout
	Timeout = time.Second
	defer func() { Timeout = old }()
	start := time.Now()
	_, err := run(t, "SIZE 3\n", nil, "sleep", "20")
	wantErr(t, err, "program timed out after 1s; nothing submitted")
	if d := time.Since(start); d > 11*time.Second {
		t.Fatalf("took %s", d)
	}
}

func TestNonZeroExit(t *testing.T) {
	_, err := run(t, "SIZE 3\n", nil, "exit", "3")
	wantErr(t, err, "program exited with code 3; nothing submitted")
}

func TestStderrStreamed(t *testing.T) {
	var buf bytes.Buffer
	got, err := run(t, "SIZE 3\n", &buf, "stderr")
	if err != nil || got != "KKK" {
		t.Fatalf("got %q, %v", got, err)
	}
	if !strings.Contains(buf.String(), "hello from stderr") {
		t.Fatalf("stderr = %q", buf.String())
	}
}

func TestMissingCommand(t *testing.T) {
	_, err := Run(context.Background(), []string{"hth-no-such-command-xyz"}, "", &bytes.Buffer{})
	wantErr(t, err, `could not start "hth-no-such-command-xyz"`)
}

func TestNoAnswerLine(t *testing.T) {
	_, err := run(t, "", nil, "sleep", "0")
	wantErr(t, err, "no answer line on stdout")
}

func TestProgramIgnoringStdin(t *testing.T) {
	script := "SIZE 3\n" + strings.Repeat("# filler line to overflow the pipe buffer\n", 20000)
	got, err := run(t, script, nil, "no-read")
	if err != nil || got != "KKK" {
		t.Fatalf("got %q, %v", got, err)
	}
}
