package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoad 测试从 YAML 文件加载配置
func TestLoad(t *testing.T) {
	// 创建临时配置文件
	content := []byte(`
app:
  name: "test-app"
  port: 9090
  env: "testing"
db:
  host: "127.0.0.1"
  port: 3306
  user: "testuser"
  password: "testpass"
  database: "testdb"
redis:
  host: "127.0.0.1"
  port: 6379
log:
  level: "debug"
  format: "json"
`)
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, content, 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// 验证各字段
	if cfg.App.Name != "test-app" {
		t.Errorf("App.Name = %q, want %q", cfg.App.Name, "test-app")
	}
	if cfg.App.Port != 9090 {
		t.Errorf("App.Port = %d, want %d", cfg.App.Port, 9090)
	}
	if cfg.DB.Host != "127.0.0.1" {
		t.Errorf("DB.Host = %q, want %q", cfg.DB.Host, "127.0.0.1")
	}
	if cfg.DB.User != "testuser" {
		t.Errorf("DB.User = %q, want %q", cfg.DB.User, "testuser")
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Log.Level = %q, want %q", cfg.Log.Level, "debug")
	}
	if cfg.Log.Format != "json" {
		t.Errorf("Log.Format = %q, want %q", cfg.Log.Format, "json")
	}
}

// TestLoadEnvOverride 测试环境变量覆盖 YAML 配置
func TestLoadEnvOverride(t *testing.T) {
	content := []byte(`
app:
  name: "yaml-name"
  port: 8080
  env: "development"
log:
  level: "info"
  format: "text"
`)
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, content, 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	// 设置环境变量覆盖
	t.Setenv("APP_NAME", "env-name")
	t.Setenv("APP_PORT", "3000")
	t.Setenv("LOG_LEVEL", "error")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// 验证环境变量覆盖生效
	if cfg.App.Name != "env-name" {
		t.Errorf("App.Name = %q, want %q (env override)", cfg.App.Name, "env-name")
	}
	if cfg.App.Port != 3000 {
		t.Errorf("App.Port = %d, want %d (env override)", cfg.App.Port, 3000)
	}
	if cfg.Log.Level != "error" {
		t.Errorf("Log.Level = %q, want %q (env override)", cfg.Log.Level, "error")
	}
	// 未设置环境变量的字段保持 YAML 值
	if cfg.App.Env != "development" {
		t.Errorf("App.Env = %q, want %q (no env override)", cfg.App.Env, "development")
	}
}

// TestLoadFileNotFound 测试配置文件不存在的错误处理
func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Load() expected error for nonexistent file, got nil")
	}
}

// TestLoadInvalidYAML 测试无效 YAML 格式的错误处理
func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "bad.yaml")
	if err := os.WriteFile(cfgPath, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Error("Load() expected error for invalid YAML, got nil")
	}
}

// TestDBDSN 测试 DSN 生成
func TestDBDSN(t *testing.T) {
	db := &DBConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "pass",
		Database: "mydb",
	}
	expected := "root:pass@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
	if got := db.DSN(); got != expected {
		t.Errorf("DSN() = %q, want %q", got, expected)
	}
}

// TestRedisAddr 测试 Redis 地址生成
func TestRedisAddr(t *testing.T) {
	r := &RedisConfig{Host: "127.0.0.1", Port: 6379}
	expected := "127.0.0.1:6379"
	if got := r.Addr(); got != expected {
		t.Errorf("Addr() = %q, want %q", got, expected)
	}
}

// TestAppEnv 测试环境判断方法
func TestAppEnv(t *testing.T) {
	tests := []struct {
		env    string
		isDev  bool
		isProd bool
	}{
		{"development", true, false},
		{"production", false, true},
		{"staging", false, false},
	}
	for _, tt := range tests {
		app := &AppConfig{Env: tt.env}
		if got := app.IsDev(); got != tt.isDev {
			t.Errorf("IsDev(%q) = %v, want %v", tt.env, got, tt.isDev)
		}
		if got := app.IsProd(); got != tt.isProd {
			t.Errorf("IsProd(%q) = %v, want %v", tt.env, got, tt.isProd)
		}
	}
}

// BenchmarkLoad 基准测试：配置加载性能
func BenchmarkLoad(b *testing.B) {
	content := []byte(`
app:
  name: "bench-app"
  port: 8080
  env: "production"
log:
  level: "info"
  format: "json"
`)
	tmpDir := b.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, content, 0644); err != nil {
		b.Fatalf("failed to write temp config: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Load(cfgPath)
	}
}
