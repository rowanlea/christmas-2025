package main

import (
	"fmt"
	"io"
	"os"
)

var version = "dev"

const usage = `usage:
  hth login <team>
  hth run <puzzle> -- <command…>
  hth test <dir> -- <command…>
  hth show <file> -- <command…>
  hth --version
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	case "--version", "version":
		fmt.Fprintln(stdout, version)
		return 0
	case "login":
		return notImplemented(stderr, "login")
	case "run", "test", "show":
		if _, _, ok := splitProgram(args[1:]); !ok {
			fmt.Fprintf(stderr, "hth %s: separate the program command with --\n%s", args[0], usage)
			return 2
		}
		return notImplemented(stderr, args[0])
	case "viewer":
		return notImplemented(stderr, "viewer")
	default:
		fmt.Fprintf(stderr, "hth: unknown command %q\n%s", args[0], usage)
		return 2
	}
}

// splitProgram splits arguments at the first "--" into hth's own arguments
// and the program command.
func splitProgram(args []string) (own, program []string, ok bool) {
	for i, a := range args {
		if a == "--" {
			return args[:i], args[i+1:], true
		}
	}
	return args, nil, false
}

func notImplemented(stderr io.Writer, name string) int {
	fmt.Fprintf(stderr, "hth %s: not implemented\n", name)
	return 1
}
