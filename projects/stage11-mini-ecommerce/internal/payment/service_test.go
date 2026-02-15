package payment

import (
	"context"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/infra"
)

func TestCreateCallback(t *testing.T) {
	idGen, _ := infra.NewSnowflake(1)
	s := NewService(idGen)
	p, err := s.Create(context.Background(), "ord-1", 100)
	if err != nil {
		t.Fatalf("create payment failed: %v", err)
	}
	updated, ok, err := s.Callback(context.Background(), p.ID, true)
	if err != nil || !ok || !updated.Success {
		t.Fatalf("callback failed")
	}
}
