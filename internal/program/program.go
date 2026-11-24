// Package program runs a team's program against a puzzle script.
package program

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// Timeout is the time limit for one run of the team's program.
var Timeout = 10 * time.Second

// Run starts argv, feeds it script on stdin and returns the single answer
// line from stdout. The program's stderr is streamed to stderr. Checking the
// answer against SIZE is left to tree.Validate.
func Run(ctx context.Context, argv []string, script string, stderr io.Writer) (string, error) {
	if len(argv) == 0 {
		return "", errors.New("no program command given")
	}
	runCtx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, argv[0], argv[1:]...)
	cmd.WaitDelay = time.Second
	cmd.Stderr = stderr
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("could not start %q: %w", argv[0], err)
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("could not start %q: %w", argv[0], err)
	}
	// A program may exit without reading its input, so write errors
	// (a broken pipe) are deliberately ignored.
	go func() {
		_, _ = io.WriteString(stdin, script)
		_ = stdin.Close()
	}()

	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("program timed out after %s; nothing submitted", Timeout)
	}
	var exitErr *exec.ExitError
	if errors.As(waitErr, &exitErr) {
		return "", fmt.Errorf("program exited with code %d; nothing submitted", exitErr.ExitCode())
	}
	if waitErr != nil && !errors.Is(waitErr, exec.ErrWaitDelay) {
		return "", fmt.Errorf("program failed: %w; nothing submitted", waitErr)
	}

	var lines []string
	for _, l := range strings.Split(stdout.String(), "\n") {
		l = strings.TrimRight(l, " \t\r")
		if l != "" {
			lines = append(lines, l)
		}
	}
	switch len(lines) {
	case 0:
		return "", errors.New("no answer line on stdout")
	case 1:
		return lines[0], nil
	}
	return "", fmt.Errorf("expected 1 answer line, got %d (print debugging to stderr)", len(lines))
}
