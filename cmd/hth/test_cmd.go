package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"hth/internal/examples"
	"hth/internal/program"
	"hth/internal/tree"
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
	for _, ex := range exs {
		if msg := checkExample(ex, argv, stderr); msg == "" {
			passed++
			fmt.Fprintf(stdout, "PASS %s\n", ex.Name)
		} else {
			fmt.Fprintf(stdout, "FAIL %s: %s\n", ex.Name, msg)
		}
	}
	fmt.Fprintf(stdout, "%d/%d passed\n", passed, len(exs))
	if passed != len(exs) {
		return 1
	}
	return 0
}

// checkExample returns "" when the program's answer matches, else the failure text.
func checkExample(ex examples.Example, argv []string, stderr io.Writer) string {
	answer, err := program.Run(context.Background(), argv, ex.Script, stderr)
	if err != nil {
		return oneLine(err.Error())
	}
	if err := tree.Validate(answer, ex.Size); err != nil {
		return oneLine(err.Error())
	}
	diff := tree.Diff(ex.Expected, answer, ex.Size)
	if len(diff) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("wrong bulbs at")
	for _, d := range diff {
		fmt.Fprintf(&b, " (%d,%d)", d[0], d[1])
	}
	return b.String()
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
