package main

import (
	"context"
	"fmt"
	"io"

	"hth/internal/api"
	"hth/internal/program"
	"hth/internal/tree"
)

// runCmd fetches a live challenge, runs the team's program on its script,
// submits the rendered answer and prints the score.
func runCmd(args []string, stdout, stderr io.Writer) int {
	own, argv, _ := splitProgram(args)
	if len(own) != 1 || len(argv) == 0 {
		fmt.Fprintln(stderr, "hth: usage: hth run <puzzle> -- <command…>")
		return 1
	}
	if err := runPuzzle(own[0], argv, stdout, stderr); err != nil {
		fmt.Fprintf(stderr, "hth: %s\n", oneLine(err.Error()))
		return 1
	}
	return 0
}

func runPuzzle(puzzle string, argv []string, stdout, stderr io.Writer) error {
	cfg, err := api.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.Team == "" || cfg.Secret == "" {
		return fmt.Errorf("not logged in; run hth login <team>")
	}
	client := api.NewClient(cfg)
	ch, err := client.FetchChallenge(puzzle)
	if err != nil {
		return err
	}
	answer, err := program.Run(context.Background(), argv, ch.Script, stderr)
	if err != nil {
		return err
	}
	if err := tree.Validate(answer, ch.Size); err != nil {
		return err
	}
	img, err := tree.Render(answer, ch.Size)
	if err != nil {
		return err
	}
	score, err := client.Submit(ch.ID, img)
	if err != nil {
		return err
	}
	pct := 0
	if score.BulbCount > 0 {
		pct = 100 * score.Correct / score.BulbCount
	}
	fmt.Fprintf(stdout, "Puzzle %s: %d/%d correct (%d%%), %d points (best %d)\n",
		score.Puzzle, score.Correct, score.BulbCount, pct, score.Points, score.BestPoints)
	if score.Unlocked != nil {
		fmt.Fprintf(stdout, "Unlocked %s\n", *score.Unlocked)
	}
	return nil
}
