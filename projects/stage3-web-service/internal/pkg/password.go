package pkg

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

var ErrInvalidPasswordHash = errors.New("invalid password hash")

// HashPassword 对密码做哈希。
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	_, _ = h.Write(salt)
	_, _ = h.Write([]byte(password))
	digest := h.Sum(nil)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedDigest := base64.RawStdEncoding.EncodeToString(digest)
	return encodedSalt + ":" + encodedDigest, nil
}

// VerifyPassword 校验密码。
func VerifyPassword(password, hashed string) error {
	parts := strings.Split(hashed, ":")
	if len(parts) != 2 {
		return ErrInvalidPasswordHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return ErrInvalidPasswordHash
	}
	expectedDigest, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return ErrInvalidPasswordHash
	}

	h := sha256.New()
	_, _ = h.Write(salt)
	_, _ = h.Write([]byte(password))
	actualDigest := h.Sum(nil)

	if !hmac.Equal(expectedDigest, actualDigest) {
		return errors.New("password mismatch")
	}
	return nil
}
