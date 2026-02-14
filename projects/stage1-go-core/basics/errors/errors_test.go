package errors

import (
	stdErrors "errors"
	"testing"
)

func TestWrapNil(t *testing.T) {
	if err := Wrap("load", nil); err != nil {
		t.Fatalf("Wrap(nil) = %v, want nil", err)
	}
}

func TestWrapAndUnwrap(t *testing.T) {
	err := Wrap("load-user", ErrInvalidInput)
	if err == nil {
		t.Fatal("Wrap() returned nil")
	}

	if !stdErrors.Is(err, ErrInvalidInput) {
		t.Fatalf("errors.Is() = false, want true")
	}

	if got := err.Error(); got != "load-user: invalid input" {
		t.Fatalf("Error() = %q, want %q", got, "load-user: invalid input")
	}
}

func TestIsNotFound(t *testing.T) {
	err := Wrap("query", ErrNotFound)
	if !IsNotFound(err) {
		t.Fatal("IsNotFound() = false, want true")
	}
}

func BenchmarkWrap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Wrap("benchmark", ErrInvalidInput)
	}
}
