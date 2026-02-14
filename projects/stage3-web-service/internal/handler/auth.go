package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/model"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/pkg"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/service"
)

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthHandler 认证处理器。
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器。
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register 处理注册请求。
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, ErrCodeBadRequest, "method not allowed")
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid request body")
		return
	}

	user, err := h.authService.Register(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, pkg.ErrInvalidUsername), errors.Is(err, pkg.ErrInvalidEmail), errors.Is(err, pkg.ErrInvalidPassword):
			Error(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
		case errors.Is(err, model.ErrUserExists):
			Error(w, http.StatusConflict, ErrCodeUserExists, err.Error())
		default:
			Error(w, http.StatusInternalServerError, ErrCodeInternal, "internal server error")
		}
		return
	}

	Success(w, map[string]any{"id": user.ID, "username": user.Username, "email": user.Email, "role": user.Role})
}

// Login 处理登录请求。
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, ErrCodeBadRequest, "method not allowed")
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid request body")
		return
	}

	accessToken, refreshToken, user, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, pkg.ErrInvalidUsername), errors.Is(err, pkg.ErrInvalidPassword):
			Error(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
		case errors.Is(err, model.ErrWrongPassword):
			Error(w, http.StatusUnauthorized, ErrCodeWrongPassword, err.Error())
		default:
			Error(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		}
		return
	}

	Success(w, map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// Refresh 刷新 access token。
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, ErrCodeBadRequest, "method not allowed")
		return
	}

	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid request body")
		return
	}

	accessToken, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		Error(w, http.StatusUnauthorized, ErrCodeUnauthorized, "invalid refresh token")
		return
	}

	Success(w, map[string]any{"access_token": accessToken})
}
