// Command fake is a configurable stand-in for a team's program, used in tests.
package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	mode := ""
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	arg := ""
	if len(os.Args) > 2 {
		arg = os.Args[2]
	}
	switch mode {
	case "echo-k":
		echoK(readScript())
	case "count":
		count(readScript())
	case "print":
		fmt.Println(arg)
	case "crlf":
		fmt.Print("KKK \r\n")
	case "debug-stdout":
		fmt.Print("debugging\nKKK\n")
	case "sleep":
		n, _ := strconv.Atoi(arg)
		time.Sleep(time.Duration(n) * time.Second)
	case "exit":
		n, _ := strconv.Atoi(arg)
		os.Exit(n)
	case "stderr":
		fmt.Fprintln(os.Stderr, "hello from stderr")
		fmt.Println("KKK")
	case "no-read":
		fmt.Println("KKK")
	case "slow-after":
		n, _ := strconv.Atoi(arg)
		script := readScript()
		if len(strings.Split(strings.TrimRight(script, "\n"), "\n")) > n {
			time.Sleep(20 * time.Second)
		}
		count(script)
	case "sleep-ms":
		n, _ := strconv.Atoi(arg)
		script := readScript()
		time.Sleep(time.Duration(n) * time.Millisecond)
		echoK(script)
	case "fail-lines":
		n, _ := strconv.Atoi(arg)
		script := readScript()
		if len(strings.Split(strings.TrimRight(script, "\n"), "\n")) == n {
			os.Exit(1)
		}
		count(script)
	case "first-call-k":
		script := readScript()
		if _, err := os.Stat(arg); err != nil {
			os.WriteFile(arg, nil, 0o644)
			echoK(script)
			return
		}
		count(script)
	case "log":
		script := readScript()
		f, err := os.OpenFile(arg, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			fmt.Fprintln(f, time.Now().Format(time.RFC3339Nano))
			f.Close()
		}
		count(script)
	default:
		fmt.Fprintln(os.Stderr, "unknown mode", mode)
		os.Exit(2)
	}
}

func readScript() string {
	b, _ := io.ReadAll(os.Stdin)
	return string(b)
}

func bulbCount(script string) int {
	for _, l := range strings.Split(script, "\n") {
		f := strings.Fields(l)
		if len(f) == 2 && f[0] == "SIZE" {
			n, _ := strconv.Atoi(f[1])
			return n * (n + 1) / 2
		}
	}
	return 0
}

func echoK(script string) {
	fmt.Println(strings.Repeat("K", bulbCount(script)))
}

func count(script string) {
	total := bulbCount(script)
	if total == 0 {
		fmt.Println()
		return
	}
	instr := 0
	for _, l := range strings.Split(script, "\n") {
		f := strings.Fields(l)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		switch f[0] {
		case "SIZE", "REPEAT", "END":
			continue
		}
		instr++
	}
	b := []byte(strings.Repeat("K", total))
	b[instr%total] = 'R'
	fmt.Println(string(b))
}
