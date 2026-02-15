package order

import (
	"context"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/infra"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/inventory"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/payment"
)

func TestCreatePayCancel(t *testing.T) {
	orderIDGen, _ := infra.NewSnowflake(1)
	payIDGen, _ := infra.NewSnowflake(2)
	inv := inventory.NewService()
	_ = inv.Seed(context.Background(), 1, 10)
	pay := payment.NewService(payIDGen)
	s := NewService(orderIDGen, inv, pay)
	items := []model.OrderItem{{ProductID: 1, Quantity: 2, Price: 50}}
	o, err := s.Create(context.Background(), 10, items)
	if err != nil {
		t.Fatalf("create order failed: %v", err)
	}
	if _, err = s.Pay(context.Background(), o.ID); err != nil {
		t.Fatalf("pay failed: %v", err)
	}
	if err = s.Cancel(context.Background(), o.ID); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
}
