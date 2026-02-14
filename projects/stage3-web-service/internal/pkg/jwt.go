package pkg

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

// Claims JWT 声明。
type Claims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Type      string `json:"type"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

// JWTManager 管理 token 签发与校验。
type JWTManager struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTManager 创建 JWT 管理器。
func NewJWTManager(secret string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	if accessExpiry <= 0 {
		accessExpiry = time.Hour
	}
	if refreshExpiry <= 0 {
		refreshExpiry = 24 * time.Hour
	}
	return &JWTManager{
		secretKey:     []byte(secret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateTokenPair 生成 access token 与 refresh token。
func (m *JWTManager) GenerateTokenPair(userID int64, username, role string) (string, string, error) {
	accessToken, err := m.signToken(userID, username, role, "access", m.accessExpiry)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := m.signToken(userID, username, role, "refresh", m.refreshExpiry)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// ParseToken 解析并校验 token。
func (m *JWTManager) ParseToken(tokenStr string) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	headerPayload := parts[0] + "." + parts[1]
	wantSig := sign(headerPayload, m.secretKey)
	if !hmac.Equal([]byte(parts[2]), []byte(wantSig)) {
		return nil, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims := &Claims{}
	if err = json.Unmarshal(payload, claims); err != nil {
		return nil, ErrInvalidToken
	}
	if time.Now().Unix() >= claims.ExpiresAt {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// RefreshToken 使用 refresh token 刷新 access token。
func (m *JWTManager) RefreshToken(refreshToken string) (string, error) {
	claims, err := m.ParseToken(refreshToken)
	if err != nil {
		return "", err
	}
	if claims.Type != "refresh" {
		return "", ErrInvalidToken
	}
	return m.signToken(claims.UserID, claims.Username, claims.Role, "access", m.accessExpiry)
}

func (m *JWTManager) signToken(userID int64, username, role, tokenType string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		Type:      tokenType,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(expiry).Unix(),
	}

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	headerPayload := fmt.Sprintf("%s.%s", header, payload)
	signature := sign(headerPayload, m.secretKey)
	return headerPayload + "." + signature, nil
}

func sign(content string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	_, _ = h.Write([]byte(content))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
