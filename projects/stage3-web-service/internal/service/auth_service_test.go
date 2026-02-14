package service

import (
	"context"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/model"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/pkg"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/repository"
)

func TestAuthServiceRegisterLogin(t *testing.T) {
	repo := repository.NewInMemoryUserRepo()
	jwtMgr := pkg.NewJWTManager("secret", time.Minute, time.Hour)
	svc := NewAuthService(repo, jwtMgr)

	user, err := svc.Register(context.Background(), "alice", "alice@example.com", "123456")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.ID <= 0 {
		t.Fatalf("user.ID = %d, want > 0", user.ID)
	}

	access, refresh, loginUser, err := svc.Login(context.Background(), "alice", "123456")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatal("Login() returned empty token")
	}
	if loginUser.Username != "alice" {
		t.Fatalf("Login user = %q, want alice", loginUser.Username)
	}
}

func TestAuthServiceWrongPassword(t *testing.T) {
	repo := repository.NewInMemoryUserRepo()
	jwtMgr := pkg.NewJWTManager("secret", time.Minute, time.Hour)
	svc := NewAuthService(repo, jwtMgr)
	_, _ = svc.Register(context.Background(), "alice", "alice@example.com", "123456")

	_, _, _, err := svc.Login(context.Background(), "alice", "wrong")
	if err != model.ErrWrongPassword {
		t.Fatalf("Login() error = %v, want %v", err, model.ErrWrongPassword)
	}
}
