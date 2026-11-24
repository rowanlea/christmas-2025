package script

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

const (
	passCap      = 100
	candidateCap = 100_000
)

// Frame is one animation point: a cut-down script that stops after one step.
type Frame struct {
	Script        string
	SkippedBefore int64
}

// Plan is the ordered frames plus passes skipped after the last frame.
type Plan struct {
	Frames       []Frame
	SkippedAtEnd int64
}

type pathEl struct {
	idx  int
	pass int64
}

type candidate struct {
	path []pathEl
	skip int64
}

// Sample picks up to max frames for hth run and hth test.
func Sample(s Script, max int) Plan {
	w := &walker{passCap: passCap, candCap: candidateCap}
	w.walk(s.Body, nil)
	cands := w.cands
	if max <= 0 || len(cands) == 0 {
		return Plan{SkippedAtEnd: w.pending}
	}
	chosen := choose(cands, max)
	plan := Plan{SkippedAtEnd: w.pending}
	for _, i := range chosen {
		plan.Frames = append(plan.Frames, Frame{Script: cutDown(s.Size, s.Body, cands[i].path), SkippedBefore: cands[i].skip})
	}
	return plan
}

// Every keeps every step with no pass cap.
func Every(s Script) Plan {
	w := &walker{passCap: -1, candCap: -1}
	w.walk(s.Body, nil)
	plan := Plan{}
	for _, c := range w.cands {
		plan.Frames = append(plan.Frames, Frame{Script: cutDown(s.Size, s.Body, c.path)})
	}
	return plan
}

type walker struct {
	passCap int64 // negative: no cap
	candCap int   // negative: no cap
	cands   []candidate
	pending int64
}

func (w *walker) capped() bool { return w.candCap >= 0 && len(w.cands) >= w.candCap }

func (w *walker) walk(nodes []Node, path []pathEl) {
	for i, n := range nodes {
		if !n.Loop {
			if w.capped() {
				continue
			}
			p := append(append([]pathEl(nil), path...), pathEl{idx: i})
			w.cands = append(w.cands, candidate{path: p, skip: w.pending})
			w.pending = 0
			continue
		}
		passes := n.Count
		if w.passCap >= 0 && passes > w.passCap {
			passes = w.passCap
		}
		for k := int64(0); k < passes; k++ {
			if w.capped() {
				w.pending = satAdd(w.pending, satMul(passes-k, skips(n.Body, w.passCap)))
				break
			}
			w.walk(n.Body, append(path, pathEl{idx: i, pass: k}))
		}
		w.pending = satAdd(w.pending, n.Count-passes)
	}
}

// skips returns the passes a capped walk of nodes would skip, arithmetically.
func skips(nodes []Node, cap int64) int64 {
	var total int64
	for _, n := range nodes {
		p := n.Count
		if cap >= 0 && p > cap {
			p = cap
		}
		total = satAdd(total, satAdd(n.Count-p, satMul(p, skips(n.Body, cap))))
	}
	return total
}

func satAdd(a, b int64) int64 {
	if a > math.MaxInt64-b {
		return math.MaxInt64
	}
	return a + b
}

func satMul(a, b int64) int64 {
	if a == 0 || b == 0 {
		return 0
	}
	if a > math.MaxInt64/b {
		return math.MaxInt64
	}
	return a * b
}

// choose returns ascending candidate indices: evenly spread, always the first,
// the last, and every candidate that follows a skip, trimmed to max where the
// forced ones leave room.
func choose(cands []candidate, max int) []int {
	n := len(cands)
	protected := map[int]bool{0: true, n - 1: true}
	for i, c := range cands {
		if c.skip > 0 {
			protected[i] = true
		}
	}
	set := map[int]bool{}
	for i := range protected {
		set[i] = true
	}
	if max >= n {
		for i := 0; i < n; i++ {
			set[i] = true
		}
	} else if max > 1 {
		for j := 0; j < max; j++ {
			set[int(math.Round(float64(j)*float64(n-1)/float64(max-1)))] = true
		}
	}
	var out, loose []int
	for i := range set {
		out = append(out, i)
		if !protected[i] {
			loose = append(loose, i)
		}
	}
	sort.Ints(out)
	sort.Ints(loose)
	if excess := len(out) - max; excess > 0 && len(loose) > 0 {
		if excess > len(loose) {
			excess = len(loose)
		}
		drop := map[int]bool{}
		for j := 0; j < excess; j++ {
			drop[loose[int((float64(j)+0.5)*float64(len(loose))/float64(excess))]] = true
		}
		kept := out[:0]
		for _, i := range out {
			if !drop[i] {
				kept = append(kept, i)
			}
		}
		out = kept
	}
	return out
}

func cutDown(size int, body []Node, path []pathEl) string {
	var b strings.Builder
	fmt.Fprintf(&b, "SIZE %d\n", size)
	cut(&b, body, path)
	return b.String()
}

func cut(b *strings.Builder, body []Node, path []pathEl) {
	el := path[0]
	emitFull(b, body[:el.idx])
	n := body[el.idx]
	if !n.Loop {
		b.WriteString(n.Line + "\n")
		return
	}
	if el.pass > 0 {
		emitFull(b, []Node{{Loop: true, Count: el.pass, Body: n.Body}})
	}
	cut(b, n.Body, path[1:])
}

func emitFull(b *strings.Builder, nodes []Node) {
	for _, n := range nodes {
		if !n.Loop {
			b.WriteString(n.Line + "\n")
			continue
		}
		fmt.Fprintf(b, "REPEAT %d\n", n.Count)
		emitFull(b, n.Body)
		b.WriteString("END\n")
	}
}
