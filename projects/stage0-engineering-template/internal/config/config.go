// Package config 负责加载和管理应用配置
// 支持 YAML 文件加载 + 环境变量覆盖
package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config 应用配置根结构
type Config struct {
	App   AppConfig   `yaml:"app"`
	DB    DBConfig    `yaml:"db"`
	Redis RedisConfig `yaml:"redis"`
	Log   LogConfig   `yaml:"log"`
}

// AppConfig 应用基本配置
type AppConfig struct {
	Name string `yaml:"name" env:"APP_NAME"`
	Port int    `yaml:"port" env:"APP_PORT"`
	Env  string `yaml:"env"  env:"APP_ENV"`
}

// DBConfig 数据库连接配置
type DBConfig struct {
	Host            string `yaml:"host"              env:"DB_HOST"`
	Port            int    `yaml:"port"              env:"DB_PORT"`
	User            string `yaml:"user"              env:"DB_USER"`
	Password        string `yaml:"password"          env:"DB_PASSWORD"`
	Database        string `yaml:"database"          env:"DB_NAME"`
	MaxOpenConns    int    `yaml:"max_open_conns"    env:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int    `yaml:"max_idle_conns"    env:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME"`
}

// RedisConfig Redis 连接配置
type RedisConfig struct {
	Host     string `yaml:"host"     env:"REDIS_HOST"`
	Port     int    `yaml:"port"     env:"REDIS_PORT"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int    `yaml:"db"       env:"REDIS_DB"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"  env:"LOG_LEVEL"`
	Format string `yaml:"format" env:"LOG_FORMAT"`
}

// Load 从 YAML 文件加载配置，然后用环境变量覆盖
// 加载优先级：YAML 文件 < 环境变量
func Load(path string) (*Config, error) {
	cfg := &Config{}

	// 1. 读取 YAML 配置文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 2. 环境变量覆盖（通过 env tag 反射读取）
	overrideFromEnv(cfg)

	return cfg, nil
}

// overrideFromEnv 递归遍历结构体，读取 env tag 对应的环境变量覆盖字段值
func overrideFromEnv(cfg interface{}) {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// 递归处理嵌套结构体
		if fieldVal.Kind() == reflect.Struct {
			overrideFromEnv(fieldVal.Addr().Interface())
			continue
		}

		// 读取 env tag
		envKey := field.Tag.Get("env")
		if envKey == "" {
			continue
		}

		envVal, ok := os.LookupEnv(envKey)
		if !ok {
			continue
		}

		// 根据字段类型设置值
		switch fieldVal.Kind() {
		case reflect.String:
			fieldVal.SetString(envVal)
		case reflect.Int:
			if intVal, err := strconv.Atoi(envVal); err == nil {
				fieldVal.SetInt(int64(intVal))
			}
		}
	}
}

// Addr 返回数据库连接地址
func (c *DBConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// DSN 返回数据库 DSN 连接字符串
func (c *DBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Database)
}

// Addr 返回 Redis 连接地址
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsDev 判断是否为开发环境
func (c *AppConfig) IsDev() bool {
	return c.Env == "development"
}

// IsProd 判断是否为生产环境
func (c *AppConfig) IsProd() bool {
	return c.Env == "production"
}
