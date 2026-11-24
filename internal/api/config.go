package api

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

const DefaultServer = "https://lscc.lifeatcodurance.com/"

type Config struct {
	Server string `json:"server"`
	Team   string `json:"team"`
	Secret string `json:"secret"`
}

func configPath() (string, error) {
	if p := os.Getenv("HTH_CONFIG"); p != "" {
		return p, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "hth", "config.json"), nil
}

// LoadConfig reads the saved config. A missing file yields an empty config
// with the default server.
func LoadConfig() (Config, error) {
	cfg := Config{Server: DefaultServer}
	path, err := configPath()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{Server: DefaultServer}, err
	}
	if cfg.Server == "" {
		cfg.Server = DefaultServer
	}
	return cfg, nil
}

func SaveConfig(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}
