package product

import (
	"context"
	"testing"
)

func TestCreateGet(t *testing.T) {
	s := NewService()
	p, err := s.Create(context.Background(), "p1", 100)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	got, ok, err := s.Get(context.Background(), p.ID)
	if err != nil || !ok || got.Name != "p1" {
		t.Fatalf("unexpected get result")
	}
}
