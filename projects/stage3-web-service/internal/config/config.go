package config

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Config 应用配置。
type Config struct {
	App AppConfig `yaml:"app"`
	JWT JWTConfig `yaml:"jwt"`
	Log LogConfig `yaml:"log"`
}

// AppConfig 应用配置。
type AppConfig struct {
	Name string `yaml:"name" env:"APP_NAME"`
	Port int    `yaml:"port" env:"APP_PORT"`
	Env  string `yaml:"env" env:"APP_ENV"`
}

// JWTConfig JWT 配置。
type JWTConfig struct {
	Secret               string `yaml:"secret" env:"JWT_SECRET"`
	AccessExpirySeconds  int    `yaml:"access_expiry_seconds" env:"JWT_ACCESS_EXPIRY_SECONDS"`
	RefreshExpirySeconds int    `yaml:"refresh_expiry_seconds" env:"JWT_REFRESH_EXPIRY_SECONDS"`
}

// LogConfig 日志配置。
type LogConfig struct {
	Level  string `yaml:"level" env:"LOG_LEVEL"`
	Format string `yaml:"format" env:"LOG_FORMAT"`
}

// Load 加载配置并应用环境变量覆盖。
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	cfg := &Config{}
	if err = parseSimpleYAML(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config failed: %w", err)
	}

	overrideFromEnv(cfg)
	if cfg.App.Port <= 0 {
		cfg.App.Port = 8081
	}
	if cfg.JWT.AccessExpirySeconds <= 0 {
		cfg.JWT.AccessExpirySeconds = 3600
	}
	if cfg.JWT.RefreshExpirySeconds <= 0 {
		cfg.JWT.RefreshExpirySeconds = 86400
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "stage3-secret"
	}

	return cfg, nil
}

func parseSimpleYAML(data []byte, cfg *Config) error {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	section := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, ":") {
			section = strings.TrimSuffix(line, ":")
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"")

		switch section {
		case "app":
			switch key {
			case "name":
				cfg.App.Name = value
			case "port":
				if n, err := strconv.Atoi(value); err == nil {
					cfg.App.Port = n
				}
			case "env":
				cfg.App.Env = value
			}
		case "jwt":
			switch key {
			case "secret":
				cfg.JWT.Secret = value
			case "access_expiry_seconds":
				if n, err := strconv.Atoi(value); err == nil {
					cfg.JWT.AccessExpirySeconds = n
				}
			case "refresh_expiry_seconds":
				if n, err := strconv.Atoi(value); err == nil {
					cfg.JWT.RefreshExpirySeconds = n
				}
			}
		case "log":
			switch key {
			case "level":
				cfg.Log.Level = value
			case "format":
				cfg.Log.Format = value
			}
		}
	}

	return scanner.Err()
}

func overrideFromEnv(target any) {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		if fieldValue.Kind() == reflect.Struct {
			overrideFromEnv(fieldValue.Addr().Interface())
			continue
		}

		envKey := field.Tag.Get("env")
		if envKey == "" {
			continue
		}
		raw, ok := os.LookupEnv(envKey)
		if !ok {
			continue
		}

		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(raw)
		case reflect.Int:
			if n, err := strconv.Atoi(raw); err == nil {
				fieldValue.SetInt(int64(n))
			}
		}
	}
}
