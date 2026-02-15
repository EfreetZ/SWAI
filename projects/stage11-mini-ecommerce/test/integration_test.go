package test

import (
	"context"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/app"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
)

func TestIntegration(t *testing.T) {
	a, err := app.New()
	if err != nil {
		t.Fatalf("new app failed: %v", err)
	}
	p, _ := a.ProductSvc.Create(context.Background(), "p1", 100)
	_ = a.InventorySvc.Seed(context.Background(), p.ID, 20)
	items := []model.OrderItem{{ProductID: p.ID, Quantity: 2, Price: 100}}
	o, _, err := a.CreateOrderAndPay(context.Background(), 1, items)
	if err != nil {
		t.Fatalf("flow failed: %v", err)
	}
	if o.Status != model.OrderPaid {
		t.Fatalf("unexpected status: %d", o.Status)
	}
}
