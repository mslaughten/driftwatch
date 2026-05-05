package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "driftwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	content := `
watch_paths:
  - /etc/myapp/config.yaml
  - /etc/nginx/nginx.conf
poll_interval: 60s
webhook:
  url: https://hooks.example.com/notify
  timeout: 5s
log_level: debug
`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.WatchPaths) != 2 {
		t.Errorf("expected 2 watch paths, got %d", len(cfg.WatchPaths))
	}
	if cfg.PollInterval != 60*time.Second {
		t.Errorf("expected 60s poll interval, got %s", cfg.PollInterval)
	}
	if cfg.Webhook.URL != "https://hooks.example.com/notify" {
		t.Errorf("unexpected webhook URL: %s", cfg.Webhook.URL)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log_level debug, got %s", cfg.LogLevel)
	}
}

func TestLoad_MissingWatchPaths(t *testing.T) {
	content := `
webhook:
  url: https://hooks.example.com/notify
`
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing watch_paths")
	}
}

func TestLoad_MissingWebhookURL(t *testing.T) {
	content := `
watch_paths:
  - /etc/app/config.yaml
`
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing webhook URL")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("expected default poll interval 30s, got %s", cfg.PollInterval)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default log level info, got %s", cfg.LogLevel)
	}
}
