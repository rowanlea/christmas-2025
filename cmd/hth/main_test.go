package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"hth/internal/api"
)

func TestLoginWithSecretSkipsServer(t *testing.T) {
	t.Setenv("HTH_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	code, _, errs := runCapture("login", "team-07", "--secret", "abc", "--server", "http://127.0.0.1:1/")
	if code != 0 {
		t.Fatalf("code %d stderr %q", code, errs)
	}
	cfg, err := api.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Team != "team-07" || cfg.Secret != "abc" || cfg.Server != "http://127.0.0.1:1/" {
		t.Fatalf("cfg %+v", cfg)
	}
}

func TestLoginRegistersWithoutSecret(t *testing.T) {
	t.Setenv("HTH_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/register/team-07" {
			t.Errorf("path %s", r.URL.Path)
		}
		w.Write([]byte(`"zzz"`))
	}))
	defer srv.Close()
	code, out, errs := runCapture("login", "team-07", "--server", srv.URL)
	if code != 0 {
		t.Fatalf("code %d stderr %q", code, errs)
	}
	want := "Logged in as team-07. Secret: zzz\nShare this secret with your pair: hth login team-07 --secret zzz\n"
	if out != want {
		t.Fatalf("out %q", out)
	}
}

func runCapture(args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestNoArgsPrintsUsage(t *testing.T) {
	code, _, stderr := runCapture()
	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	for _, want := range []string{"hth login <team>", "hth run <puzzle> -- <command…>", "hth test <dir> -- <command…>", "hth show <file> -- <command…>"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("usage missing %q in %q", want, stderr)
		}
	}
}

func TestUnknownCommand(t *testing.T) {
	if code, _, _ := runCapture("frobnicate"); code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
}

func TestRunWithoutSeparator(t *testing.T) {
	code, _, stderr := runCapture("run", "3.2", "python")
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "hth: usage: hth run <puzzle> -- <command…>") {
		t.Errorf("stderr should carry the run usage line: %q", stderr)
	}
}

func TestVersion(t *testing.T) {
	code, stdout, _ := runCapture("--version")
	if code != 0 || !strings.Contains(stdout, "dev") {
		t.Fatalf("code=%d stdout=%q", code, stdout)
	}
}
