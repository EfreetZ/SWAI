package inventory

import (
	"context"
	"errors"
	"sync"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
)

var ErrInsufficientStock = errors.New("insufficient stock")

// Service 库存服务。
type Service struct {
	mu    sync.Mutex
	items map[int64]model.Inventory
}

func NewService() *Service {
	return &Service{items: make(map[int64]model.Inventory)}
}

func (s *Service) Seed(ctx context.Context, productID int64, stock int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[productID] = model.Inventory{ProductID: productID, Stock: stock, Locked: 0}
	return nil
}

func (s *Service) Deduct(ctx context.Context, productID int64, quantity int) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	inv := s.items[productID]
	if inv.Stock-inv.Locked < int64(quantity) {
		return ErrInsufficientStock
	}
	inv.Locked += int64(quantity)
	s.items[productID] = inv
	return nil
}

func (s *Service) Confirm(ctx context.Context, productID int64, quantity int) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	inv := s.items[productID]
	inv.Locked -= int64(quantity)
	inv.Stock -= int64(quantity)
	if inv.Locked < 0 {
		inv.Locked = 0
	}
	if inv.Stock < 0 {
		inv.Stock = 0
	}
	s.items[productID] = inv
	return nil
}

func (s *Service) Rollback(ctx context.Context, productID int64, quantity int) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	inv := s.items[productID]
	inv.Locked -= int64(quantity)
	if inv.Locked < 0 {
		inv.Locked = 0
	}
	s.items[productID] = inv
	return nil
}

func (s *Service) Get(ctx context.Context, productID int64) (model.Inventory, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return model.Inventory{}, false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.items[productID]
	return inv, ok, nil
}
