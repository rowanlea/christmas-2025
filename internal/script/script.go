// Package script parses the structure of a puzzle script and plans animation
// frames. It never interprets instructions: only SIZE, REPEAT/END, comments
// and blank lines are known; every other line is an opaque step.
package script

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

const maxRepeat = 1_000_000_000

// Node is one step (Line set) or one loop (Loop true, Count passes of Body).
type Node struct {
	Line  string
	Count int64
	Body  []Node
	Loop  bool
}

// Script is a parsed script.
type Script struct {
	Size int
	Body []Node
}

// Parse reads script text. Line numbers in errors are 1-based physical lines.
func Parse(text string) (Script, error) {
	var s Script
	type open struct {
		count int64
		line  int
		body  []Node
	}
	var stack []open
	cur := []Node{}
	sawSize := false
	lines := strings.Split(text, "\n")
	for i, raw := range lines {
		n := i + 1
		line := strings.Trim(raw, " \t\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if !sawSize {
			if fields[0] != "SIZE" {
				return Script{}, fmt.Errorf("line %d: missing SIZE", n)
			}
			if len(fields) != 2 {
				return Script{}, fmt.Errorf("line %d: bad SIZE", n)
			}
			size, err := strconv.Atoi(fields[1])
			if err != nil || size < 1 {
				return Script{}, fmt.Errorf("line %d: bad SIZE", n)
			}
			s.Size = size
			sawSize = true
			continue
		}
		switch fields[0] {
		case "REPEAT":
			var count int64
			var err error
			if len(fields) == 2 {
				count, err = strconv.ParseInt(fields[1], 10, 64)
			}
			if len(fields) != 2 || err != nil || count < 0 || count > maxRepeat {
				return Script{}, fmt.Errorf("line %d: bad REPEAT count", n)
			}
			stack = append(stack, open{count: count, line: n, body: cur})
			cur = []Node{}
		case "END":
			if len(stack) == 0 {
				return Script{}, fmt.Errorf("line %d: unmatched END", n)
			}
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			cur = append(top.body, Node{Loop: true, Count: top.count, Body: cur})
		default:
			cur = append(cur, Node{Line: line})
		}
	}
	if !sawSize {
		return Script{}, fmt.Errorf("line 1: missing SIZE")
	}
	if len(stack) > 0 {
		return Script{}, fmt.Errorf("line %d: unclosed REPEAT", stack[len(stack)-1].line)
	}
	s.Body = cur
	return s, nil
}

// Steps returns the total executed steps, counted without walking any loop.
func (s Script) Steps() *big.Int { return steps(s.Body) }

func steps(nodes []Node) *big.Int {
	total := new(big.Int)
	for _, n := range nodes {
		if !n.Loop {
			total.Add(total, big.NewInt(1))
			continue
		}
		inner := steps(n.Body)
		inner.Mul(inner, big.NewInt(n.Count))
		total.Add(total, inner)
	}
	return total
}
