package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadRepoExampleFile(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "configs", "example.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":8080" {
		t.Fatalf("addr: got %q", cfg.Addr)
	}
	if len(cfg.Targets) != 0 {
		t.Fatalf("targets: got %d", len(cfg.Targets))
	}
}

func TestLoadExampleShape(t *testing.T) {
	path := writeConfig(t, `
addr: ":8080"
interval: 30s
timeout: 5s
targets: []
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":8080" {
		t.Fatalf("addr: got %q", cfg.Addr)
	}
	if cfg.Interval != 30*time.Second {
		t.Fatalf("interval: got %s", cfg.Interval)
	}
	if cfg.Timeout != 5*time.Second {
		t.Fatalf("timeout: got %s", cfg.Timeout)
	}
	if cfg.WebhookURL != "" {
		t.Fatalf("webhook: got %q", cfg.WebhookURL)
	}
	if len(cfg.Targets) != 0 {
		t.Fatalf("targets: got %d", len(cfg.Targets))
	}
}

func TestLoadTargets(t *testing.T) {
	path := writeConfig(t, `
addr: ":8080"
interval: 30s
timeout: 5s
webhook_url: https://example.com/hook
targets:
  - name: api
    url: https://api.example.com/health
    expect_status: 200
  - name: web
    url: http://example.com
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WebhookURL != "https://example.com/hook" {
		t.Fatalf("webhook: got %q", cfg.WebhookURL)
	}
	if len(cfg.Targets) != 2 {
		t.Fatalf("targets: got %d", len(cfg.Targets))
	}
	if cfg.Targets[0].Name != "api" || cfg.Targets[0].ExpectStatus != 200 {
		t.Fatalf("api target: %+v", cfg.Targets[0])
	}
	if cfg.Targets[1].Name != "web" || cfg.Targets[1].ExpectStatus != 0 {
		t.Fatalf("web target: %+v", cfg.Targets[1])
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	path := writeConfig(t, `
addr: ":8080"
interval: 30s
timeout: 5s
webhook_url: https://example.com/hook
targets: []
`)
	t.Setenv("ADDR", ":9090")
	t.Setenv("WEBHOOK_URL", "https://hooks.example.com/override")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":9090" {
		t.Fatalf("addr: got %q", cfg.Addr)
	}
	if cfg.WebhookURL != "https://hooks.example.com/override" {
		t.Fatalf("webhook: got %q", cfg.WebhookURL)
	}
}

func TestLoadRejectsEmptyName(t *testing.T) {
	path := writeConfig(t, `
addr: ":8080"
interval: 30s
timeout: 5s
targets:
  - name: ""
    url: https://example.com
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadRejectsNonHTTPURL(t *testing.T) {
	path := writeConfig(t, `
addr: ":8080"
interval: 30s
timeout: 5s
targets:
  - name: ftp
    url: ftp://example.com/file
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadRejectsDuplicateNames(t *testing.T) {
	path := writeConfig(t, `
addr: ":8080"
interval: 30s
timeout: 5s
targets:
  - name: api
    url: https://a.example.com
  - name: api
    url: https://b.example.com
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error")
	}
}
