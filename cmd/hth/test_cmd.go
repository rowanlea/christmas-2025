package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"hth/internal/animate"
	"hth/internal/examples"
	"hth/internal/program"
	"hth/internal/script"
	"hth/internal/tree"
	"hth/internal/viewer"
)

// testCmd runs the team's program against every example in a directory.
func testCmd(args []string, stdout, stderr io.Writer) int {
	own, argv, ok := splitProgram(args)
	if !ok || len(argv) == 0 || len(own) != 1 {
		fmt.Fprintf(stderr, "hth test: expected <dir> -- <command…>\n%s", usage)
		return 2
	}
	exs, err := examples.Load(own[0])
	if err != nil {
		fmt.Fprintf(stderr, "hth test: %v\n", err)
		return 1
	}
	passed := 0
	answers := make([]string, len(exs))
	summaries := make([]string, len(exs))
	for i, ex := range exs {
		msg, answer := checkExample(ex, argv, stderr)
		answers[i] = answer
		if msg == "" {
			passed++
			summaries[i] = "PASS"
			fmt.Fprintf(stdout, "PASS %s\n", ex.Name)
		} else {
			summaries[i] = "FAIL: " + msg
			fmt.Fprintf(stdout, "FAIL %s: %s\n", ex.Name, msg)
		}
	}
	fmt.Fprintf(stdout, "%d/%d passed\n", passed, len(exs))
	animateExamples(exs, answers, summaries, argv, stdout, stderr)
	if passed != len(exs) {
		return 1
	}
	return 0
}

// animateExamples is best effort and never changes the exit code. An example
// whose program gave no valid answer has no final tree and is not animated.
func animateExamples(exs []examples.Example, answers, summaries []string, argv []string, stdout, stderr io.Writer) {
	plans := make([]script.Plan, len(exs))
	cards := make([]viewer.Card, len(exs))
	for i, ex := range exs {
		cards[i] = viewer.Card{
			ID: ex.Name, Title: ex.Name, Size: ex.Size, Answer: answers[i], Expected: ex.Expected,
			Summary: summaries[i], IntervalMs: 100,
		}
		if answers[i] == "" {
			continue
		}
		s, err := script.Parse(ex.Script)
		if err != nil {
			fmt.Fprintf(stderr, "hth: gif: %s: %s\n", ex.Name, oneLine(err.Error()))
			answers[i] = ""
			continue
		}
		plans[i] = script.Sample(s, maxFrames)
		cards[i].FrameCount = len(plans[i].Frames)
	}
	sess := openSession("test", cards, stderr)
	if sess != nil {
		defer sess.Close()
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	for i, ex := range exs {
		if answers[i] == "" {
			if sess != nil {
				sess.Done(ex.Name, 0)
			}
			continue
		}
		fmt.Fprintf(stdout, "Animating %s: %d frames…\n", ex.Name, len(plans[i].Frames))
		gifPath := "hth-test-" + ex.Name + ".gif"
		if _, err := animate.Animate(ctx, argv, plans[i], ex.Size, answers[i], ex.Name, sess, gifPath); err != nil {
			if ctx.Err() != nil {
				return
			}
			fmt.Fprintf(stderr, "hth: gif: %s\n", oneLine(err.Error()))
			continue
		}
		fmt.Fprintf(stdout, "Saved %s\n", gifPath)
	}
}

// checkExample returns "" when the program's answer matches, else the failure
// text. The answer is returned whenever it is valid for the example's size.
func checkExample(ex examples.Example, argv []string, stderr io.Writer) (msg, answer string) {
	answer, err := program.Run(context.Background(), argv, ex.Script, stderr)
	if err != nil {
		return oneLine(err.Error()), ""
	}
	if err := tree.Validate(answer, ex.Size); err != nil {
		return oneLine(err.Error()), ""
	}
	diff := tree.Diff(ex.Expected, answer, ex.Size)
	if len(diff) == 0 {
		return "", answer
	}
	var b strings.Builder
	b.WriteString("wrong bulbs at")
	for _, d := range diff {
		fmt.Fprintf(&b, " (%d,%d)", d[0], d[1])
	}
	return b.String(), answer
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
