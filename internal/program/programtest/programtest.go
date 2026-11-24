// Package programtest builds the fake team program for tests.
package programtest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

var (
	once     sync.Once
	path     string
	buildErr error
)

// FakePath builds the fake program once per test binary and returns its path.
func FakePath(t testing.TB) string {
	t.Helper()
	once.Do(func() {
		dir, err := os.MkdirTemp("", "hth-fake-")
		if err != nil {
			buildErr = err
			return
		}
		name := "fake"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		path = filepath.Join(dir, name)
		_, file, _, _ := runtime.Caller(0)
		src := filepath.Join(filepath.Dir(file), "..", "testdata", "fake")
		out, err := exec.Command("go", "build", "-o", path, src).CombinedOutput()
		if err != nil {
			buildErr = fmt.Errorf("%w\n%s", err, out)
		}
	})
	if buildErr != nil {
		t.Fatalf("building fake program: %v", buildErr)
	}
	return path
}
