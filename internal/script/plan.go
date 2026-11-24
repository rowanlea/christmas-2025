package script

import (
	"fmt"
	"math"
	"math/big"
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

// Sample picks up to limit frames for hth run and hth test. It never returns
// more than limit frames.
func Sample(s Script, limit int) Plan {
	w := &walker{passCap: passCap, candCap: candidateCap}
	w.walk(s.Body, nil)
	cands := w.cands
	if limit <= 0 || len(cands) == 0 {
		return Plan{SkippedAtEnd: w.pending}
	}
	plan := Plan{SkippedAtEnd: w.pending}
	prev := -1
	for _, i := range choose(cands, limit) {
		var skipped int64
		for _, c := range cands[prev+1 : i+1] {
			skipped = satAdd(skipped, c.skip)
		}
		prev = i
		plan.Frames = append(plan.Frames, Frame{Script: cutDown(s.Size, s.Body, cands[i].path), SkippedBefore: skipped})
	}
	return plan
}

// Every keeps every step with no pass cap. It returns an empty Plan when the
// script has more than 500 steps.
func Every(s Script) Plan {
	if s.Steps().Cmp(big.NewInt(500)) > 0 {
		return Plan{}
	}
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
		if !hasSteps(n.Body) {
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
		if !n.Loop || !hasSteps(n.Body) {
			continue
		}
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

// hasSteps reports whether executing nodes runs at least one step.
func hasSteps(nodes []Node) bool {
	for _, n := range nodes {
		if !n.Loop || (n.Count > 0 && hasSteps(n.Body)) {
			return true
		}
	}
	return false
}

// choose returns at most limit ascending candidate indices: evenly spread,
// including the first, the last and, where room allows, every candidate that
// follows a skip.
func choose(cands []candidate, limit int) []int {
	n := len(cands)
	if limit < 2 {
		return []int{0}
	}
	if limit >= n {
		all := make([]int, n)
		for i := range all {
			all[i] = i
		}
		return all
	}
	var forced []int
	for i, c := range cands {
		if i == 0 || i == n-1 || c.skip > 0 {
			forced = append(forced, i)
		}
	}
	if len(forced) >= limit {
		out := []int{forced[0]}
		out = append(out, spread(forced[1:len(forced)-1], limit-2)...)
		return append(out, forced[len(forced)-1])
	}
	set := map[int]bool{}
	for _, i := range forced {
		set[i] = true
	}
	for j := 0; j < limit; j++ {
		set[int(math.Round(float64(j)*float64(n-1)/float64(limit-1)))] = true
	}
	var out, loose []int
	for i := range set {
		out = append(out, i)
	}
	sort.Ints(out)
	isForced := map[int]bool{}
	for _, i := range forced {
		isForced[i] = true
	}
	for _, i := range out {
		if !isForced[i] {
			loose = append(loose, i)
		}
	}
	if excess := len(out) - limit; excess > 0 {
		keep := map[int]bool{}
		for _, i := range spread(loose, len(loose)-excess) {
			keep[i] = true
		}
		kept := out[:0]
		for _, i := range out {
			if isForced[i] || keep[i] {
				kept = append(kept, i)
			}
		}
		out = kept
	}
	return out
}

// spread picks k items of list evenly, keeping order.
func spread(list []int, k int) []int {
	if k >= len(list) {
		return list
	}
	var out []int
	for j := 0; j < k; j++ {
		out = append(out, list[int((float64(j)+0.5)*float64(len(list))/float64(k))])
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
