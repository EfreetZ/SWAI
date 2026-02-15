package order

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/infra"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/inventory"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/payment"
)

var ErrOrderNotFound = errors.New("order not found")

// Service 订单服务。
type Service struct {
	mu        sync.RWMutex
	orders    map[string]model.Order
	idGen     *infra.Snowflake
	inventory *inventory.Service
	payment   *payment.Service
}

func NewService(idGen *infra.Snowflake, inv *inventory.Service, pay *payment.Service) *Service {
	return &Service{orders: make(map[string]model.Order), idGen: idGen, inventory: inv, payment: pay}
}

func (s *Service) Create(ctx context.Context, userID int64, items []model.OrderItem) (model.Order, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return model.Order{}, err
	}
	var total int64
	for _, item := range items {
		if err := s.inventory.Deduct(ctx, item.ProductID, item.Quantity); err != nil {
			return model.Order{}, err
		}
		total += int64(item.Quantity) * item.Price
	}
	id, err := s.idGen.NextID()
	if err != nil {
		for _, item := range items {
			_ = s.inventory.Rollback(ctx, item.ProductID, item.Quantity)
		}
		return model.Order{}, err
	}
	now := time.Now()
	o := model.Order{ID: formatOrderID(id), UserID: userID, Items: items, TotalPrice: total, Status: model.OrderPending, CreatedAt: now, UpdatedAt: now}
	s.mu.Lock()
	s.orders[o.ID] = o
	s.mu.Unlock()
	return o, nil
}

func (s *Service) Pay(ctx context.Context, orderID string) (model.Payment, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return model.Payment{}, err
	}
	s.mu.RLock()
	o, ok := s.orders[orderID]
	s.mu.RUnlock()
	if !ok {
		return model.Payment{}, ErrOrderNotFound
	}
	p, err := s.payment.Create(ctx, orderID, o.TotalPrice)
	if err != nil {
		for _, item := range o.Items {
			_ = s.inventory.Rollback(ctx, item.ProductID, item.Quantity)
		}
		return model.Payment{}, err
	}
	_, _, err = s.payment.Callback(ctx, p.ID, true)
	if err != nil {
		return model.Payment{}, err
	}
	for _, item := range o.Items {
		_ = s.inventory.Confirm(ctx, item.ProductID, item.Quantity)
	}
	o.Status = model.OrderPaid
	o.UpdatedAt = time.Now()
	s.mu.Lock()
	s.orders[o.ID] = o
	s.mu.Unlock()
	return p, nil
}

func (s *Service) Cancel(ctx context.Context, orderID string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	o, ok := s.orders[orderID]
	if !ok {
		return ErrOrderNotFound
	}
	if o.Status == model.OrderPending {
		for _, item := range o.Items {
			_ = s.inventory.Rollback(ctx, item.ProductID, item.Quantity)
		}
	}
	o.Status = model.OrderCancelled
	o.UpdatedAt = time.Now()
	s.orders[o.ID] = o
	return nil
}

func (s *Service) Get(ctx context.Context, orderID string) (model.Order, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return model.Order{}, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[orderID]
	return o, ok, nil
}

func formatOrderID(v int64) string {
	return "ord-" + int64ToString(v)
}

func int64ToString(v int64) string {
	if v == 0 {
		return "0"
	}
	buf := make([]byte, 0, 20)
	for v > 0 {
		buf = append([]byte{byte('0' + v%10)}, buf...)
		v /= 10
	}
	return string(buf)
}
