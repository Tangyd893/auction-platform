package model

import "time"

// Item 拍品
type Item struct {
	ID           int64     `db:"id" json:"id"`
	Title        string    `db:"title" json:"title"`
	Description  string    `db:"description" json:"description"`
	ImageURL     string    `db:"image_url" json:"image_url"`
	StartPrice   int64     `db:"start_price" json:"start_price"`     // 起拍价（分）
	CurrentPrice int64     `db:"current_price" json:"current_price"` // 当前价
	ReservePrice int64     `db:"reserve_price" json:"reserve_price"` // 保留价
	BidIncrement int64     `db:"bid_increment" json:"bid_increment"`   // 最低加价幅度
	SellerID     int64     `db:"seller_id" json:"seller_id"`
	Status       string    `db:"status" json:"status"`
	StartTime    time.Time `db:"start_time" json:"start_time"`
	EndTime      time.Time `db:"end_time" json:"end_time"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type ItemStatus string

const (
	ItemStatusDraft   ItemStatus = "draft"
	ItemStatusListed  ItemStatus = "listed"
	ItemStatusSold    ItemStatus = "sold"
	ItemStatusUnsold  ItemStatus = "unsold"
	ItemStatusCancel  ItemStatus = "cancelled"
)
