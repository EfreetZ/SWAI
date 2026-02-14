package pkg

import (
	"errors"
	"strings"
)

var (
	ErrInvalidUsername = errors.New("invalid username")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidEmail    = errors.New("invalid email")
)

// ValidateRegisterParams 校验注册参数。
func ValidateRegisterParams(username, email, password string) error {
	if strings.TrimSpace(username) == "" {
		return ErrInvalidUsername
	}
	if !strings.Contains(email, "@") {
		return ErrInvalidEmail
	}
	if len(password) < 6 {
		return ErrInvalidPassword
	}
	return nil
}

// ValidateLoginParams 校验登录参数。
func ValidateLoginParams(username, password string) error {
	if strings.TrimSpace(username) == "" {
		return ErrInvalidUsername
	}
	if strings.TrimSpace(password) == "" {
		return ErrInvalidPassword
	}
	return nil
}
