// Package animate builds an animation by running the team's program on each
// cut-down frame script, streaming frames to the viewer and writing a GIF.
package animate

import (
	"context"
	"io"
	"strconv"
	"sync"

	"hth/internal/art"
	"hth/internal/program"
	"hth/internal/script"
	"hth/internal/tree"
	"hth/internal/viewer"
)

const workers = 8

// Animate runs every frame script of plan through the program, sending each
// valid answer to sess (which may be nil) as it arrives, then writes the GIF
// from the frames that succeeded, in plan order. The passes skipped before a
// failed frame carry to the next kept frame in the GIF. It returns the kept
// frame lines in plan order, without the final tree.
func Animate(ctx context.Context, argv []string, plan script.Plan, size int, final string, card string, sess *viewer.Session, gifPath string) ([]string, error) {
	answers := make([]string, len(plan.Frames))
	jobs := make(chan int, len(plan.Frames))
	for i := range plan.Frames {
		jobs <- i
	}
	close(jobs)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				if ctx.Err() != nil {
					return
				}
				line, err := program.Run(ctx, argv, plan.Frames[i].Script, io.Discard)
				if err != nil || tree.Validate(line, size) != nil {
					continue
				}
				answers[i] = line
				if sess != nil {
					sess.Frame(viewer.FrameEvent{Card: card, Index: i, Line: line, SkippedBefore: plan.Frames[i].SkippedBefore})
				}
			}
		}()
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var kept []string
	captions := map[int]string{}
	var carried int64
	for i, f := range plan.Frames {
		carried += f.SkippedBefore
		if answers[i] == "" {
			continue
		}
		if carried > 0 {
			captions[len(kept)] = Caption(carried)
		}
		kept = append(kept, answers[i])
		carried = 0
	}
	if tail := plan.SkippedAtEnd + carried; tail > 0 {
		captions[len(kept)] = Caption(tail)
	}
	if sess != nil {
		sess.Done(card, plan.SkippedAtEnd+carried)
	}
	return kept, art.WriteGIF(gifPath, size, kept, captions, final)
}

// Caption is the text shown over a frame that follows n skipped passes, with
// thousands separators as the viewer shows them (en-GB).
func Caption(n int64) string {
	s := strconv.FormatInt(n, 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s + " passes later"
}
