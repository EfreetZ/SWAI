package limiter

import "testing"

func TestAllow(t *testing.T) {
	b := NewTokenBucket(1, 1)
	if !b.Allow() {
		t.Fatal("first token should pass")
	}
	if b.Allow() {
		t.Fatal("second token should be denied immediately")
	}
}
