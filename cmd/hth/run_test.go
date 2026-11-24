package main

import (
	"bytes"
	"encoding/json"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"hth/internal/api"
	"hth/internal/program/programtest"
	"hth/internal/viewer"
)

type fakeAPI struct {
	mu          sync.Mutex
	submissions [][]byte
	fetchStatus int
	fetchMsg    string
	submitCode  int
	script      string
	submittedAt time.Time
}

func (f *fakeAPI) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /puzzles/{puzzle}/challenges", func(w http.ResponseWriter, r *http.Request) {
		if f.fetchStatus != 0 {
			w.WriteHeader(f.fetchStatus)
			json.NewEncoder(w).Encode(map[string]string{"error": f.fetchMsg})
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]any{
			"challengeId": "c1", "puzzle": r.PathValue("puzzle"), "size": 16,
			"script": f.scriptText(), "expiresAt": "2030-01-01T00:00:00Z",
		})
	})
	mux.HandleFunc("POST /challenges/{id}/submission", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		f.mu.Lock()
		f.submissions = append(f.submissions, body)
		f.submittedAt = time.Now()
		f.mu.Unlock()
		if f.submitCode != 0 {
			w.WriteHeader(f.submitCode)
			json.NewEncoder(w).Encode(map[string]string{"error": "expired"})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"puzzle": "1.1", "correct": 136, "bulbCount": 136,
			"points": 100, "bestPoints": 100, "unlocked": "1.2",
		})
	})
	return mux
}

func setup(t *testing.T, f *fakeAPI) {
	t.Helper()
	srv := httptest.NewServer(f.handler())
	t.Cleanup(srv.Close)
	loginTo(t, srv.URL)
}

func loginTo(t *testing.T, url string) {
	t.Helper()
	t.Chdir(t.TempDir())
	t.Setenv("HTH_CONFIG", t.TempDir()+"/config.json")
	if err := api.SaveConfig(api.Config{Server: url, Team: "team-07", Secret: "s"}); err != nil {
		t.Fatal(err)
	}
}

func TestRunHappyPath(t *testing.T) {
	f := &fakeAPI{}
	setup(t, f)
	code, out, errs := runCapture("run", "1.1", "--", programtest.FakePath(t), "echo-k")
	if code != 0 {
		t.Fatalf("code = %d, stderr %q", code, errs)
	}
	for _, want := range []string{"Puzzle 1.1: 136/136 correct (100%), 100 points (best 100)", "Unlocked 1.2"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q: %q", want, out)
		}
	}
	if len(f.submissions) != 1 {
		t.Fatalf("got %d submissions, want 1", len(f.submissions))
	}
	img, err := png.Decode(bytes.NewReader(f.submissions[0]))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 1200 || b.Dy() != 1200 {
		t.Errorf("png is %dx%d, want 1200x1200", b.Dx(), b.Dy())
	}
}

func TestRunWrongLengthSubmitsNothing(t *testing.T) {
	f := &fakeAPI{}
	setup(t, f)
	code, _, errs := runCapture("run", "1.1", "--", programtest.FakePath(t), "print", "KKK")
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.HasPrefix(errs, "hth: ") {
		t.Errorf("stderr = %q", errs)
	}
	if len(f.submissions) != 0 {
		t.Errorf("got %d submissions, want none", len(f.submissions))
	}
}

func TestRunNotLoggedIn(t *testing.T) {
	t.Setenv("HTH_CONFIG", t.TempDir()+"/missing.json")
	code, _, errs := runCapture("run", "1.1", "--", programtest.FakePath(t), "echo-k")
	if code != 1 || !strings.Contains(errs, "not logged in; run hth login <team>") {
		t.Errorf("code = %d, stderr %q", code, errs)
	}
}

func TestRunLocked(t *testing.T) {
	f := &fakeAPI{fetchStatus: 403, fetchMsg: "puzzle 1.2 is locked: reach 80% on 1.1 first"}
	setup(t, f)
	code, _, errs := runCapture("run", "1.2", "--", programtest.FakePath(t), "echo-k")
	if code != 1 || !strings.Contains(errs, f.fetchMsg) {
		t.Errorf("code = %d, stderr %q", code, errs)
	}
}

func TestRunExpired(t *testing.T) {
	f := &fakeAPI{submitCode: 410}
	setup(t, f)
	code, _, errs := runCapture("run", "1.1", "--", programtest.FakePath(t), "echo-k")
	if code != 1 || !strings.Contains(errs, "expired") {
		t.Errorf("code = %d, stderr %q", code, errs)
	}
}

func TestRunServerDown(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()
	loginTo(t, url)
	code, _, errs := runCapture("run", "1.1", "--", programtest.FakePath(t), "echo-k")
	if code != 1 || !strings.Contains(errs, "could not reach") {
		t.Errorf("code = %d, stderr %q", code, errs)
	}
}

func TestRunUsage(t *testing.T) {
	for _, args := range [][]string{
		{"run", "--", "x"},
		{"run", "1.1", "--"},
		{"run", "1.1"},
		{"run"},
	} {
		code, _, errs := runCapture(args...)
		if code != 1 || !strings.Contains(errs, "hth: usage: hth run <puzzle> -- <command…>") {
			t.Errorf("%v: code = %d, stderr %q", args, code, errs)
		}
	}
}

func TestRunNotLoggedInEmptySecret(t *testing.T) {
	t.Setenv("HTH_CONFIG", t.TempDir()+"/config.json")
	if err := api.SaveConfig(api.Config{Server: "http://127.0.0.1:1/", Team: "team-07"}); err != nil {
		t.Fatal(err)
	}
	code, _, errs := runCapture("run", "1.1", "--", programtest.FakePath(t), "echo-k")
	if code != 1 || !strings.Contains(errs, "not logged in; run hth login <team>") {
		t.Errorf("code = %d, stderr %q", code, errs)
	}
}

func TestRunStreamsProgramStderr(t *testing.T) {
	setup(t, &fakeAPI{})
	_, _, errs := runCapture("run", "1.1", "--", programtest.FakePath(t), "stderr")
	if !strings.Contains(errs, "hello from stderr") {
		t.Errorf("stderr = %q", errs)
	}
}

func (f *fakeAPI) scriptText() string {
	if f.script != "" {
		return f.script
	}
	return "SIZE 16\n"
}

func TestMain(m *testing.M) {
	openViewer = func(viewer.Publication) (*viewer.Session, error) { return nil, nil }
	os.Exit(m.Run())
}

func TestRunSubmitsBeforeAnimating(t *testing.T) {
	f := &fakeAPI{script: "SIZE 16\nA\nB\nC\n"}
	setup(t, f)
	logPath := filepath.Join(t.TempDir(), "runs.log")
	code, out, errs := runCapture("run", "1.1", "--", programtest.FakePath(t), "log", logPath)
	if code != 0 {
		t.Fatalf("code = %d, stderr %q", code, errs)
	}
	for _, want := range []string{"Animating 3 frames…", "Saved hth-run-1.1.gif"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q: %q", want, out)
		}
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Fields(string(data))
	if len(lines) != 4 {
		t.Fatalf("got %d program runs, want 4", len(lines))
	}
	f.mu.Lock()
	submitted := f.submittedAt
	f.mu.Unlock()
	before := 0
	for _, l := range lines {
		started, err := time.Parse(time.RFC3339Nano, l)
		if err != nil {
			t.Fatal(err)
		}
		if started.Before(submitted) {
			before++
		}
	}
	if before != 1 {
		t.Errorf("%d runs started before the submission, want 1", before)
	}
}

func TestRunSucceedsWhenViewerFails(t *testing.T) {
	openViewer = viewer.Open
	t.Cleanup(func() { openViewer = func(viewer.Publication) (*viewer.Session, error) { return nil, nil } })
	html := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, "<html>not hth</html>")
	}))
	t.Cleanup(html.Close)
	u, _ := url.Parse(html.URL)
	t.Setenv("HTH_VIEWER_PORT", u.Port())
	setup(t, &fakeAPI{})
	code, _, errs := runCapture("run", "1.1", "--", programtest.FakePath(t), "echo-k")
	if code != 0 {
		t.Fatalf("code = %d, stderr %q", code, errs)
	}
	if !strings.Contains(errs, "hth: viewer:") {
		t.Errorf("stderr = %q", errs)
	}
	if _, err := os.Stat("hth-run-1.1.gif"); err != nil {
		t.Errorf("gif not written: %v", err)
	}
}
