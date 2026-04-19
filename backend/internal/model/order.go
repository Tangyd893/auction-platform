package model

import "time"

// Order 成交订单
type Order struct {
	ID         int64     `db:"id" json:"id"`
	ItemID     int64     `db:"item_id" json:"item_id"`
	SellerID   int64     `db:"seller_id" json:"seller_id"`
	BuyerID    int64     `db:"buyer_id" json:"buyer_id"`
	FinalPrice int64     `db:"final_price" json:"final_price"` // 最终成交价（分）
	Status     string    `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	PaidAt     *time.Time `db:"paid_at" json:"paid_at,omitempty"`
}

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)
