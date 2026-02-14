package types

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidEnv = errors.New("invalid env")

// Env 表示运行环境。
type Env string

const (
	EnvDevelopment Env = "development"
	EnvTesting     Env = "testing"
	EnvProduction  Env = "production"
)

// ParseEnv 将字符串解析为标准环境枚举。
func ParseEnv(value string) (Env, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "development", "dev":
		return EnvDevelopment, nil
	case "testing", "test":
		return EnvTesting, nil
	case "production", "prod":
		return EnvProduction, nil
	default:
		return "", ErrInvalidEnv
	}
}

// Pair 表示一对值。
type Pair[L any, R any] struct {
	Left  L
	Right R
}

// SwapPair 返回交换左右值后的 Pair。
func SwapPair[L any, R any](pair Pair[L, R]) Pair[R, L] {
	return Pair[R, L]{Left: pair.Right, Right: pair.Left}
}

// ParseIntDefault 将字符串解析为 int，空字符串时返回默认值。
func ParseIntDefault(raw string, defaultValue int) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(trimmed)
	if err != nil {
		return defaultValue, err
	}
	return value, nil
}

// Filter 返回满足 predicate 的元素集合。
func Filter[T any](items []T, predicate func(T) bool) []T {
	if predicate == nil {
		return nil
	}

	result := make([]T, 0, len(items))
	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}
