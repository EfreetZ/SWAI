package pkg

import (
	"testing"
	"time"
)

func TestJWTManagerGenerateParseRefresh(t *testing.T) {
	mgr := NewJWTManager("secret", time.Minute, 2*time.Minute)
	access, refresh, err := mgr.GenerateTokenPair(1, "alice", "admin")
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}
	claims, err := mgr.ParseToken(access)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UserID != 1 || claims.Type != "access" {
		t.Fatalf("claims unexpected: %+v", claims)
	}

	newAccess, err := mgr.RefreshToken(refresh)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if newAccess == "" {
		t.Fatal("RefreshToken() returned empty token")
	}
}

func TestJWTExpired(t *testing.T) {
	mgr := NewJWTManager("secret", time.Millisecond, time.Millisecond)
	access, _, err := mgr.GenerateTokenPair(1, "alice", "admin")
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}
	time.Sleep(3 * time.Millisecond)
	if _, err = mgr.ParseToken(access); err == nil {
		t.Fatal("ParseToken() want expired error")
	}
}
