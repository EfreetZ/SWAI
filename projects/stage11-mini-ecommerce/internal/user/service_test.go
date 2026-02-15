package user

import (
	"context"
	"testing"
)

func TestRegisterLogin(t *testing.T) {
	s := NewService()
	if _, err := s.Register(context.Background(), "u1", "p1"); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if _, err := s.Login(context.Background(), "u1", "p1"); err != nil {
		t.Fatalf("login failed: %v", err)
	}
}
