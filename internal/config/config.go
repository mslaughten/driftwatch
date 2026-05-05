package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the driftwatch daemon configuration.
type Config struct {
	WatchPaths  []string      `yaml:"watch_paths"`
	PollInterval time.Duration `yaml:"poll_interval"`
	Webhook     WebhookConfig `yaml:"webhook"`
	LogLevel    string        `yaml:"log_level"`
}

// WebhookConfig holds webhook delivery settings.
type WebhookConfig struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	Timeout time.Duration     `yaml:"timeout"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		PollInterval: 30 * time.Second,
		LogLevel:     "info",
		Webhook: WebhookConfig{
			Timeout: 10 * time.Second,
		},
	}
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config file %q: %w", path, err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate checks that required fields are present and values are sane.
func (c *Config) Validate() error {
	if len(c.WatchPaths) == 0 {
		return fmt.Errorf("watch_paths must contain at least one path")
	}
	if c.Webhook.URL == "" {
		return fmt.Errorf("webhook.url is required")
	}
	if c.PollInterval < time.Second {
		return fmt.Errorf("poll_interval must be at least 1s, got %s", c.PollInterval)
	}
	return nil
}
