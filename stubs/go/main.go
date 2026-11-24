package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	script, _ := io.ReadAll(os.Stdin)

	n := 0
	for _, raw := range strings.Split(string(script), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		n, _ = strconv.Atoi(strings.Fields(line)[1])
		break
	}

	// Start here: replace this with your emulator. Read the script above, then
	// work out the colour of every bulb. Print debugging to stderr, never stdout.
	fmt.Println(strings.Repeat("K", n*(n+1)/2))
}
