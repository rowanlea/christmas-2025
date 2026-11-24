package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"hth/internal/program/programtest"
)

type showAPI struct {
	mu        sync.Mutex
	permCode  int
	permCalls int
	uploads   [][]string
}

func (s *showAPI) start(t *testing.T) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /shows/permission", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.permCalls++
		s.mu.Unlock()
		if s.permCode != 0 {
			w.WriteHeader(s.permCode)
			json.NewEncoder(w).Encode(map[string]string{"error": "Santa comes first: get full marks on 7.1 to unlock light shows"})
			return
		}
		w.WriteHeader(204)
	})
	mux.HandleFunc("POST /shows", func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Frames []string }
		json.NewDecoder(r.Body).Decode(&body)
		s.mu.Lock()
		s.uploads = append(s.uploads, body.Frames)
		s.mu.Unlock()
		w.WriteHeader(201)
		fmt.Fprintf(w, `{"frames":%d}`, len(body.Frames))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	loginTo(t, srv.URL)
}

func writeShow(t *testing.T, name, header string, steps int) string {
	t.Helper()
	var b strings.Builder
	b.WriteString(header + "\n")
	for i := 0; i < steps; i++ {
		fmt.Fprintf(&b, "step %d\n", i)
	}
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestShowRefusedBeforeSanta(t *testing.T) {
	api := &showAPI{permCode: 403}
	api.start(t)
	file := writeShow(t, "s.txt", "SIZE 16", 3)
	logPath := filepath.Join(t.TempDir(), "runs.log")
	code, _, errs := runCapture("show", file, "--", programtest.FakePath(t), "log", logPath)
	if code != 1 || !strings.Contains(errs, "Santa comes first") {
		t.Fatalf("code = %d, stderr %q", code, errs)
	}
	if data, _ := os.ReadFile(logPath); len(data) != 0 {
		t.Errorf("program ran: %q", data)
	}
	if len(api.uploads) != 0 {
		t.Errorf("uploaded %d shows", len(api.uploads))
	}
}

func TestShowTooLong(t *testing.T) {
	api := &showAPI{}
	api.start(t)
	file := writeShow(t, "s.txt", "SIZE 16", 501)
	code, _, errs := runCapture("show", file, "--", programtest.FakePath(t), "count")
	if code != 1 || !strings.Contains(errs, "show runs 501 steps; the limit is 500") {
		t.Fatalf("code = %d, stderr %q", code, errs)
	}
	if api.permCalls != 0 {
		t.Errorf("permission check made %d calls, want none", api.permCalls)
	}
}

func TestShowWrongSize(t *testing.T) {
	api := &showAPI{}
	api.start(t)
	file := writeShow(t, "s.txt", "SIZE 4", 3)
	code, _, errs := runCapture("show", file, "--", programtest.FakePath(t), "count")
	if code != 1 || !strings.Contains(errs, "light shows use SIZE 16") {
		t.Fatalf("code = %d, stderr %q", code, errs)
	}
	if api.permCalls != 0 {
		t.Errorf("permission check made %d calls, want none", api.permCalls)
	}
}

func TestShowUploadsEveryStep(t *testing.T) {
	api := &showAPI{}
	api.start(t)
	file := writeShow(t, "my-show.txt", "SIZE 16", 12)
	code, out, errs := runCapture("show", file, "--", programtest.FakePath(t), "count")
	if code != 0 {
		t.Fatalf("code = %d, stderr %q", code, errs)
	}
	if len(api.uploads) != 1 || len(api.uploads[0]) != 13 {
		t.Fatalf("uploads = %d, frames %v", len(api.uploads), len(api.uploads))
	}
	for i, f := range api.uploads[0] {
		if len(f) != 136 {
			t.Errorf("frame %d has %d chars, want 136", i, len(f))
		}
	}
	if !strings.Contains(out, "Show uploaded: 13 frames (about 2 seconds on the big screen)") {
		t.Errorf("stdout = %q", out)
	}
	if strings.Contains(out, "were skipped") {
		t.Errorf("unexpected skip line: %q", out)
	}
	if _, err := os.Stat("hth-show-my-show.gif"); err != nil {
		t.Errorf("gif not written: %v", err)
	}
}

func TestShowFullRunFailureUploadsNothing(t *testing.T) {
	api := &showAPI{}
	api.start(t)
	file := writeShow(t, "s.txt", "SIZE 16", 3)
	code, out, errs := runCapture("show", file, "--", programtest.FakePath(t), "print", "KKK")
	if code != 1 || !strings.HasPrefix(errs, "hth: ") {
		t.Fatalf("code = %d, stderr %q", code, errs)
	}
	if len(api.uploads) != 0 || strings.Contains(out, "Animating") {
		t.Errorf("uploads %d, stdout %q", len(api.uploads), out)
	}
}
