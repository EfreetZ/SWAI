package service

import (
	"context"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/repository"
)

func TestRBACCheckerHasPermission(t *testing.T) {
	checker := NewRBACChecker(repository.NewInMemoryRoleRepo())
	ok, err := checker.HasPermission(context.Background(), "admin", "user", "delete")
	if err != nil {
		t.Fatalf("HasPermission() error = %v", err)
	}
	if !ok {
		t.Fatal("admin delete user should be allowed")
	}

	ok, err = checker.HasPermission(context.Background(), "viewer", "user", "delete")
	if err != nil {
		t.Fatalf("HasPermission() error = %v", err)
	}
	if ok {
		t.Fatal("viewer delete user should be denied")
	}
}
