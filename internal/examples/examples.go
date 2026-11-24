// Package examples loads the worked-example scripts shipped with hth.
package examples

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Example is one worked-example script with its expected answer.
type Example struct {
	Name, Script string
	Size         int
	Expected     string
}

// Load reads every *.txt file in dir, sorted by name.
func Load(dir string) ([]Example, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.txt"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no .txt examples in %s", dir)
	}
	sort.Strings(files)
	var out []Example
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		ex := Example{
			Name:   strings.TrimSuffix(filepath.Base(f), ".txt"),
			Script: string(data),
		}
		haveExpect, haveSize := false, false
		for _, line := range strings.Split(ex.Script, "\n") {
			line = strings.TrimSpace(line)
			if rest, ok := strings.CutPrefix(line, "# expect:"); ok && !haveExpect {
				ex.Expected, haveExpect = strings.TrimSpace(rest), true
			} else if rest, ok := strings.CutPrefix(line, "SIZE "); ok && !haveSize {
				n, err := strconv.Atoi(strings.TrimSpace(rest))
				if err != nil {
					return nil, fmt.Errorf("%s: bad SIZE line %q", filepath.Base(f), line)
				}
				ex.Size, haveSize = n, true
			}
		}
		if !haveExpect {
			return nil, fmt.Errorf("%s: missing \"# expect:\" line", filepath.Base(f))
		}
		if !haveSize {
			return nil, fmt.Errorf("%s: missing SIZE line", filepath.Base(f))
		}
		out = append(out, ex)
	}
	return out, nil
}
