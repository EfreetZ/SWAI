package app

import (
	"context"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
)

func TestCreateOrderAndPay(t *testing.T) {
	a, err := New()
	if err != nil {
		t.Fatalf("new app failed: %v", err)
	}
	p, err := a.ProductSvc.Create(context.Background(), "p1", 100)
	if err != nil {
		t.Fatalf("create product failed: %v", err)
	}
	_ = a.InventorySvc.Seed(context.Background(), p.ID, 100)
	items := []model.OrderItem{{ProductID: p.ID, Quantity: 2, Price: 100}}
	o, _, err := a.CreateOrderAndPay(context.Background(), 1, items)
	if err != nil {
		t.Fatalf("create and pay failed: %v", err)
	}
	if o.Status != model.OrderPaid {
		t.Fatalf("unexpected order status: %d", o.Status)
	}
}
