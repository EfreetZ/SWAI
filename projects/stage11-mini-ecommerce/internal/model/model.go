package model

import "time"

// User 用户。
type User struct {
	ID       int64
	Username string
	Password string
}

// Product 商品。
type Product struct {
	ID    int64
	Name  string
	Price int64
}

// Inventory 库存。
type Inventory struct {
	ProductID int64
	Stock     int64
	Locked    int64
}

// OrderStatus 订单状态。
type OrderStatus int

const (
	OrderPending OrderStatus = iota
	OrderPaid
	OrderCancelled
)

// OrderItem 订单项。
type OrderItem struct {
	ProductID int64
	Quantity  int
	Price     int64
}

// Order 订单。
type Order struct {
	ID         string
	UserID     int64
	Items      []OrderItem
	TotalPrice int64
	Status     OrderStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Payment 支付。
type Payment struct {
	ID        string
	OrderID   string
	Amount    int64
	Success   bool
	CreatedAt time.Time
}
