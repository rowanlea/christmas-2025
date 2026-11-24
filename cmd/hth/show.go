package main

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"hth/internal/animate"
	"hth/internal/api"
	"hth/internal/program"
	"hth/internal/script"
	"hth/internal/tree"
	"hth/internal/viewer"
)

const (
	showSize     = 16
	maxShowSteps = 500
	showFPS      = 10
)

// showCmd runs the team's program on every step of a script and uploads the
// frames as a light show.
func showCmd(args []string, stdout, stderr io.Writer) int {
	own, argv, _ := splitProgram(args)
	if len(own) != 1 || len(argv) == 0 {
		fmt.Fprintln(stderr, "hth: usage: hth show <file> -- <command…>")
		return 1
	}
	if err := showFile(own[0], argv, stdout, stderr); err != nil {
		if err != errInterrupted {
			fmt.Fprintf(stderr, "hth: %s\n", oneLine(err.Error()))
		}
		return 1
	}
	return 0
}

var errInterrupted = fmt.Errorf("interrupted")

func showFile(path string, argv []string, stdout, stderr io.Writer) error {
	text, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	s, err := script.Parse(string(text))
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if s.Size != showSize {
		return fmt.Errorf("light shows use SIZE %d", showSize)
	}
	if steps := s.Steps(); steps.Cmp(big.NewInt(maxShowSteps)) > 0 {
		return fmt.Errorf("show runs %s steps; the limit is %d", steps, maxShowSteps)
	}
	cfg, err := api.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.Team == "" || cfg.Secret == "" {
		return fmt.Errorf("not logged in; run hth login <team>")
	}
	client := api.NewClient(cfg)
	if err := client.ShowPermission(); err != nil {
		return err
	}
	answer, err := program.Run(context.Background(), argv, string(text), stderr)
	if err != nil {
		return err
	}
	if err := tree.Validate(answer, showSize); err != nil {
		return err
	}

	plan := script.Every(s)
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	card := viewer.Card{
		ID: base, Title: "Show " + base, Size: showSize, Answer: answer,
		FrameCount: len(plan.Frames), IntervalMs: 1000 / showFPS,
	}
	ctx, sess, done := animationSession("show", []viewer.Card{card}, stderr)
	defer done()
	fmt.Fprintf(stdout, "Animating %d frames…\n", len(plan.Frames))
	gifPath := "hth-show-" + base + ".gif"
	kept, err := animate.Animate(ctx, argv, plan, showSize, answer, base, sess, gifPath)
	if ctx.Err() != nil {
		return errInterrupted
	}
	if err != nil {
		fmt.Fprintf(stderr, "hth: gif: %s\n", oneLine(err.Error()))
	} else {
		fmt.Fprintf(stdout, "Saved %s\n", gifPath)
	}

	frames := kept
	if len(kept) == 0 || kept[len(kept)-1] != answer {
		if len(kept) >= maxShowSteps {
			frames = append(kept[:maxShowSteps-1:maxShowSteps-1], answer)
		} else {
			frames = append(frames, answer)
		}
	}
	if err := client.UploadShow(frames); err != nil {
		return err
	}
	if k := len(plan.Frames) - len(kept); k > 0 {
		fmt.Fprintf(stdout, "%d steps were skipped because the program failed on them\n", k)
	}
	fmt.Fprintf(stdout, "Show uploaded: %d frames (about %d seconds on the big screen)\n",
		len(frames), (len(frames)+showFPS-1)/showFPS)
	return nil
}
