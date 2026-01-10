package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	APIToken    string `json:"api_token"`
	WorkspaceID string `json:"workspace_id"`
	BaseURL     string `json:"base_url,omitempty"`
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "toggl-daily-summary", "config.json"), nil
}

func Load(path string) (Config, error) {
	if path == "" {
		defaultPath, err := DefaultPath()
		if err != nil {
			return Config{}, err
		}
		path = defaultPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func ApplyEnv(cfg *Config) {
	if cfg == nil {
		return
	}
	if v := os.Getenv("TOGGL_API_TOKEN"); v != "" {
		cfg.APIToken = v
	}
	if v := os.Getenv("TOGGL_WORKSPACE_ID"); v != "" {
		cfg.WorkspaceID = v
	}
	if v := os.Getenv("TOGGL_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
}
