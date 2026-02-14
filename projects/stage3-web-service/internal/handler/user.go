package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/model"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/repository"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/service"
)

type updateUserRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// UserHandler 用户处理器。
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器。
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// ListUsers 用户列表。
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		Error(w, http.StatusMethodNotAllowed, ErrCodeBadRequest, "method not allowed")
		return
	}
	users, err := h.userService.List(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, ErrCodeInternal, "internal server error")
		return
	}
	Success(w, users)
}

// UserByID 处理 /users/{id} 的 GET/PUT/DELETE。
func (h *UserHandler) UserByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path, "/api/v1/users/")
	if err != nil {
		Error(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid user id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getUser(w, r, id)
	case http.MethodPut:
		h.updateUser(w, r, id)
	case http.MethodDelete:
		h.deleteUser(w, r, id)
	default:
		Error(w, http.StatusMethodNotAllowed, ErrCodeBadRequest, "method not allowed")
	}
}

func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request, id int64) {
	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			Error(w, http.StatusNotFound, ErrCodeNotFound, "user not found")
			return
		}
		Error(w, http.StatusInternalServerError, ErrCodeInternal, "internal server error")
		return
	}
	Success(w, user)
}

func (h *UserHandler) updateUser(w http.ResponseWriter, r *http.Request, id int64) {
	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid request body")
		return
	}

	user := &model.User{ID: id, Email: req.Email, Role: req.Role}
	if err := h.userService.Update(r.Context(), user); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			Error(w, http.StatusNotFound, ErrCodeNotFound, "user not found")
			return
		}
		Error(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
		return
	}
	Success(w, map[string]any{"id": id, "email": req.Email, "role": req.Role})
}

func (h *UserHandler) deleteUser(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.userService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			Error(w, http.StatusNotFound, ErrCodeNotFound, "user not found")
			return
		}
		Error(w, http.StatusBadRequest, ErrCodeBadRequest, err.Error())
		return
	}
	Success(w, map[string]any{"deleted": true})
}

func parseID(path, prefix string) (int64, error) {
	raw := strings.TrimPrefix(path, prefix)
	if raw == path || raw == "" || strings.Contains(raw, "/") {
		return 0, errors.New("invalid id")
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}
