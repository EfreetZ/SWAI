package pkg

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword() returned empty hash")
	}

	if err = VerifyPassword("secret123", hash); err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if err = VerifyPassword("wrong", hash); err == nil {
		t.Fatal("VerifyPassword() want mismatch error")
	}
}

func BenchmarkHashPassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword("bench-pass")
	}
}
