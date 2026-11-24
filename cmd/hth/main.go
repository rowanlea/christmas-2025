package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"hth/internal/api"
)

var version = "dev"

const usage = `usage:
  hth login <team> [--secret S] [--server URL]
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
		return login(args[1:], stdout, stderr)
	case "test":
		return testCmd(args[1:], stdout, stderr)
	case "run", "show":
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

func login(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	fs.SetOutput(stderr)
	secret := fs.String("secret", "", "team secret (skips registration)")
	server := fs.String("server", "", "server URL")
	var teams []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return 2
		}
		args = fs.Args()
		if len(args) > 0 {
			teams = append(teams, args[0])
			args = args[1:]
		}
	}
	if len(teams) != 1 {
		fmt.Fprintf(stderr, "hth login: expected one team name\n%s", usage)
		return 2
	}
	team := teams[0]

	cfg, err := api.LoadConfig()
	if err != nil {
		fmt.Fprintf(stderr, "hth login: %v\n", err)
		return 1
	}
	if *server != "" {
		cfg.Server = *server
	}
	cfg.Team = team
	if *secret != "" {
		cfg.Secret = *secret
	} else {
		s, err := api.NewClient(cfg).Register(team)
		if err != nil {
			fmt.Fprintf(stderr, "hth login: %v\n", err)
			return 1
		}
		cfg.Secret = s
	}
	if err := api.SaveConfig(cfg); err != nil {
		fmt.Fprintf(stderr, "hth login: %v\n", err)
		return 1
	}
	if *secret == "" {
		fmt.Fprintf(stdout, "Logged in as %s. Secret: %s\n", team, cfg.Secret)
		fmt.Fprintf(stdout, "Share this secret with your pair: hth login %s --secret %s\n", team, cfg.Secret)
	} else {
		fmt.Fprintf(stdout, "Logged in as %s.\n", team)
	}
	return 0
}
