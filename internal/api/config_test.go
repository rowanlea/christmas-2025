package api

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestConfigRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "config.json")
	t.Setenv("HTH_CONFIG", path)

	want := Config{Server: "https://example.test", Team: "team-07", Secret: "s3cret"}
	if err := SaveConfig(want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("mode = %v, want 0600", info.Mode().Perm())
		}
	}
}

func TestLoadConfigMissingFileGivesDefaultServer(t *testing.T) {
	t.Setenv("HTH_CONFIG", filepath.Join(t.TempDir(), "none.json"))
	got, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got.Server != DefaultServer || got.Team != "" || got.Secret != "" {
		t.Fatalf("got %+v", got)
	}
}
