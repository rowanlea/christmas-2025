package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"time"

	"hth/internal/animate"
	"hth/internal/api"
	"hth/internal/program"
	"hth/internal/script"
	"hth/internal/tree"
	"hth/internal/viewer"
)

const maxFrames = 120

// openViewer is a variable so tests do not start a daemon or a browser.
var openViewer = viewer.Open

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
	summary := fmt.Sprintf("%d/%d correct (%d%%)", score.Correct, score.BulbCount, pct)
	animateRun(puzzle, ch.Size, answer, summary, ch.Script, argv, stdout, stderr)
	return nil
}

// animateRun is best effort: the submission has already happened, so a viewer
// or GIF problem is a warning and never changes the exit code.
func animateRun(puzzle string, size int, answer, summary, scriptText string, argv []string, stdout, stderr io.Writer) {
	s, err := script.Parse(scriptText)
	if err != nil {
		fmt.Fprintf(stderr, "hth: gif: %s\n", oneLine(err.Error()))
		return
	}
	plan := script.Sample(s, maxFrames)
	card := viewer.Card{
		ID: puzzle, Title: "Puzzle " + puzzle, Size: size, Answer: answer,
		Summary: summary, FrameCount: len(plan.Frames), IntervalMs: 100,
	}
	sess := openSession("run", []viewer.Card{card}, stderr)
	if sess != nil {
		defer sess.Close()
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	fmt.Fprintf(stdout, "Animating %d frames…\n", len(plan.Frames))
	gifPath := "hth-run-" + puzzle + ".gif"
	if _, err := animate.Animate(ctx, argv, plan, size, answer, puzzle, sess, gifPath); err != nil {
		fmt.Fprintf(stderr, "hth: gif: %s\n", oneLine(err.Error()))
		return
	}
	fmt.Fprintf(stdout, "Saved %s\n", gifPath)
}

// openSession opens a viewer publication. A failure prints a warning and
// returns nil, which every caller treats as "no viewer".
func openSession(kind string, cards []viewer.Card, stderr io.Writer) *viewer.Session {
	pub := viewer.Publication{ID: strconv.FormatInt(time.Now().UnixNano(), 36), Kind: kind, Cards: cards}
	sess, err := openViewer(pub)
	if err != nil {
		fmt.Fprintf(stderr, "hth: viewer: %s\n", oneLine(err.Error()))
		return nil
	}
	return sess
}
