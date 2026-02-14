package ratelimiter

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucketAllow(t *testing.T) {
	bucket, err := NewTokenBucket(10, 2)
	if err != nil {
		t.Fatalf("NewTokenBucket() error = %v", err)
	}

	if !bucket.Allow() {
		t.Fatal("Allow() first request should pass")
	}
	if !bucket.Allow() {
		t.Fatal("Allow() second request should pass")
	}
	if bucket.Allow() {
		t.Fatal("Allow() third request should be rejected")
	}
}

func TestTokenBucketWait(t *testing.T) {
	bucket, err := NewTokenBucket(1, 1)
	if err != nil {
		t.Fatalf("NewTokenBucket() error = %v", err)
	}
	_ = bucket.Allow()

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	if err = bucket.Wait(ctx); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
}

func TestSlidingWindowLimiter(t *testing.T) {
	limiter := NewSlidingWindowLimiter(2, 100*time.Millisecond)
	if !limiter.Allow() || !limiter.Allow() {
		t.Fatal("first two requests should pass")
	}
	if limiter.Allow() {
		t.Fatal("third request should be rejected")
	}
	time.Sleep(120 * time.Millisecond)
	if !limiter.Allow() {
		t.Fatal("request after window should pass")
	}
}

func BenchmarkTokenBucketAllow(b *testing.B) {
	bucket, err := NewTokenBucket(100000, 100000)
	if err != nil {
		b.Fatalf("NewTokenBucket() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bucket.Allow()
	}
}
