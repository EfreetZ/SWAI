package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/handler"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/pkg"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/repository"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/service"
)

func TestRequestID(t *testing.T) {
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if RequestIDFromContext(r.Context()) == "" {
			t.Fatal("request id missing in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("X-Request-ID header is empty")
	}
}

func TestAuthAndRBAC(t *testing.T) {
	jwtMgr := pkg.NewJWTManager("secret", time.Minute, time.Hour)
	access, _, err := jwtMgr.GenerateTokenPair(1, "alice", "viewer")
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	checker := service.NewRBACChecker(repository.NewInMemoryRoleRepo())
	h := Auth(jwtMgr, RBAC(checker, "user", "delete", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.Success(w, map[string]bool{"ok": true})
	})))

	r := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	r.Header.Set("Authorization", "Bearer "+access)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestRecovery(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewTextHandler(buf, nil))
	h := Recovery(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func BenchmarkAuthMiddleware(b *testing.B) {
	jwtMgr := pkg.NewJWTManager("secret", time.Minute, time.Hour)
	access, _, err := jwtMgr.GenerateTokenPair(1, "alice", "admin")
	if err != nil {
		b.Fatalf("GenerateTokenPair() error = %v", err)
	}
	h := Auth(jwtMgr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = context.Background()
		w.WriteHeader(http.StatusOK)
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+access)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
	}
}
