package structs

import (
	"errors"
	"testing"
)

func TestUserValidate(t *testing.T) {
	t.Parallel()

	valid := &User{ID: 1, Name: "alice", Email: "alice@example.com"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	invalidID := &User{Name: "alice", Email: "alice@example.com"}
	if err := invalidID.Validate(); !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("invalid ID error = %v", err)
	}

	invalidName := &User{ID: 1, Email: "alice@example.com"}
	if err := invalidName.Validate(); !errors.Is(err, ErrInvalidUserName) {
		t.Fatalf("invalid name error = %v", err)
	}

	invalidEmail := &User{ID: 1, Name: "alice", Email: "alice.example.com"}
	if err := invalidEmail.Validate(); !errors.Is(err, ErrInvalidUserEmail) {
		t.Fatalf("invalid email error = %v", err)
	}
}

func TestUserClone(t *testing.T) {
	t.Parallel()

	u := User{ID: 1, Name: "alice", Email: "alice@example.com", Tags: []string{"go"}}
	clone := u.Clone()
	clone.Tags[0] = "changed"

	if u.Tags[0] != "go" {
		t.Fatalf("Clone() modified original tags")
	}
}

func TestUserAddTag(t *testing.T) {
	t.Parallel()

	u := &User{ID: 1, Name: "alice", Email: "alice@example.com"}
	u.AddTag("backend")
	u.AddTag("backend")
	u.AddTag(" ")

	if len(u.Tags) != 1 {
		t.Fatalf("len(Tags) = %d, want 1", len(u.Tags))
	}
}

func BenchmarkUserValidate(b *testing.B) {
	u := &User{ID: 1, Name: "alice", Email: "alice@example.com"}
	for i := 0; i < b.N; i++ {
		_ = u.Validate()
	}
}
