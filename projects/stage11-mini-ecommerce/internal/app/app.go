package app

import (
	"context"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/infra"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/inventory"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/monitoring"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/order"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/payment"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/product"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/user"
)

// App 聚合电商服务。
type App struct {
	UserSvc      *user.Service
	ProductSvc   *product.Service
	InventorySvc *inventory.Service
	OrderSvc     *order.Service
	PaymentSvc   *payment.Service
	Metrics      *monitoring.Metrics
}

// New 创建应用。
func New() (*App, error) {
	idGen1, err := infra.NewSnowflake(1)
	if err != nil {
		return nil, err
	}
	idGen2, err := infra.NewSnowflake(2)
	if err != nil {
		return nil, err
	}
	userSvc := user.NewService()
	productSvc := product.NewService()
	invSvc := inventory.NewService()
	paySvc := payment.NewService(idGen2)
	orderSvc := order.NewService(idGen1, invSvc, paySvc)
	return &App{UserSvc: userSvc, ProductSvc: productSvc, InventorySvc: invSvc, OrderSvc: orderSvc, PaymentSvc: paySvc, Metrics: &monitoring.Metrics{}}, nil
}

// CreateOrderAndPay 下单并支付（简化 Saga）。
func (a *App) CreateOrderAndPay(ctx context.Context, userID int64, items []model.OrderItem) (model.Order, model.Payment, error) {
	a.Metrics.IncCreated()
	o, err := a.OrderSvc.Create(ctx, userID, items)
	if err != nil {
		a.Metrics.IncFailed()
		return model.Order{}, model.Payment{}, err
	}
	p, err := a.OrderSvc.Pay(ctx, o.ID)
	if err != nil {
		a.Metrics.IncFailed()
		return model.Order{}, model.Payment{}, err
	}
	a.Metrics.IncPaid()
	o2, _, _ := a.OrderSvc.Get(ctx, o.ID)
	return o2, p, nil
}
