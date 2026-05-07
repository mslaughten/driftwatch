package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the full driftwatch configuration.
type Config struct {
	WatchPaths    []string      `yaml:"watch_paths"`
	WebhookURL    string        `yaml:"webhook_url"`
	Interval      time.Duration `yaml:"interval"`
	SnapshotFile  string        `yaml:"snapshot_file"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Interval:     30 * time.Second,
		SnapshotFile: "/var/lib/driftwatch/snapshot.json",
	}
}

// Load reads a YAML config file and merges it with defaults.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if len(cfg.WatchPaths) == 0 {
		return nil, errors.New("config: watch_paths must not be empty")
	}
	if cfg.WebhookURL == "" {
		return nil, errors.New("config: webhook_url must not be empty")
	}
	return &cfg, nil
}
