package model

import "time"

// Bid 出价
type Bid struct {
	ID        int64     `db:"id" json:"id"`
	ItemID    int64     `db:"item_id" json:"item_id"`
	BuyerID   int64     `db:"buyer_id" json:"buyer_id"`
	Amount    int64     `db:"amount" json:"amount"`      // 出价金额（分）
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type BidStatus string

const (
	BidStatusActive  BidStatus = "active"
	BidStatusOutbid  BidStatus = "outbid"
	BidStatusWinning BidStatus = "winning"
)
