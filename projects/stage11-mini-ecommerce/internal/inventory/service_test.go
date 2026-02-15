package inventory

import (
	"context"
	"testing"
)

func TestDeductConfirmRollback(t *testing.T) {
	s := NewService()
	_ = s.Seed(context.Background(), 1, 10)
	if err := s.Deduct(context.Background(), 1, 3); err != nil {
		t.Fatalf("deduct failed: %v", err)
	}
	if err := s.Confirm(context.Background(), 1, 2); err != nil {
		t.Fatalf("confirm failed: %v", err)
	}
	inv, ok, _ := s.Get(context.Background(), 1)
	if !ok || inv.Stock != 8 {
		t.Fatalf("unexpected inventory: %+v", inv)
	}
	if err := s.Rollback(context.Background(), 1, 1); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}
}
