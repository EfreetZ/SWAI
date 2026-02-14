package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	content := []byte(`
app:
  name: "svc"
  port: 8088
  env: "testing"
jwt:
  secret: "abc"
  access_expiry_seconds: 100
  refresh_expiry_seconds: 200
log:
  level: "debug"
  format: "json"
`)
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Port != 8088 {
		t.Fatalf("App.Port = %d, want 8088", cfg.App.Port)
	}
	if cfg.JWT.Secret != "abc" {
		t.Fatalf("JWT.Secret = %q, want abc", cfg.JWT.Secret)
	}
}

func TestLoadEnvOverride(t *testing.T) {
	content := []byte("app:\n  port: 8081\n")
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("APP_PORT", "9000")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.App.Port != 9000 {
		t.Fatalf("App.Port = %d, want 9000", cfg.App.Port)
	}
}
