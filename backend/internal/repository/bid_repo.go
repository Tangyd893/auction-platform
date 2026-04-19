package repository

import (
	"database/sql"

	"auction-platform/internal/model"
)

type BidRepository struct {
	db *sql.DB
}

func NewBidRepository(db *sql.DB) *BidRepository {
	return &BidRepository{db: db}
}

func (r *BidRepository) Create(bid *model.Bid) error {
	query := `INSERT INTO bids (item_id, buyer_id, amount, status) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	return r.db.QueryRow(query, bid.ItemID, bid.BuyerID, bid.Amount, bid.Status).Scan(&bid.ID, &bid.CreatedAt)
}

func (r *BidRepository) GetByID(id int64) (*model.Bid, error) {
	query := `SELECT id, item_id, buyer_id, amount, status, created_at FROM bids WHERE id = $1`
	bid := &model.Bid{}
	err := r.db.QueryRow(query, id).Scan(&bid.ID, &bid.ItemID, &bid.BuyerID, &bid.Amount, &bid.Status, &bid.CreatedAt)
	if err != nil {
		return nil, err
	}
	return bid, nil
}

func (r *BidRepository) GetHighestBid(itemID int64) (*model.Bid, error) {
	query := `SELECT id, item_id, buyer_id, amount, status, created_at FROM bids WHERE item_id = $1 ORDER BY amount DESC LIMIT 1`
	bid := &model.Bid{}
	err := r.db.QueryRow(query, itemID).Scan(&bid.ID, &bid.ItemID, &bid.BuyerID, &bid.Amount, &bid.Status, &bid.CreatedAt)
	if err != nil {
		return nil, err
	}
	return bid, nil
}

func (r *BidRepository) ListByItemID(itemID int64) ([]*model.Bid, error) {
	query := `SELECT id, item_id, buyer_id, amount, status, created_at FROM bids WHERE item_id = $1 ORDER BY amount DESC`
	rows, err := r.db.Query(query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []*model.Bid
	for rows.Next() {
		bid := &model.Bid{}
		if err := rows.Scan(&bid.ID, &bid.ItemID, &bid.BuyerID, &bid.Amount, &bid.Status, &bid.CreatedAt); err != nil {
			return nil, err
		}
		bids = append(bids, bid)
	}
	return bids, nil
}

func (r *BidRepository) ListByBuyerID(buyerID int64) ([]*model.Bid, error) {
	query := `SELECT id, item_id, buyer_id, amount, status, created_at FROM bids WHERE buyer_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, buyerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []*model.Bid
	for rows.Next() {
		bid := &model.Bid{}
		if err := rows.Scan(&bid.ID, &bid.ItemID, &bid.BuyerID, &bid.Amount, &bid.Status, &bid.CreatedAt); err != nil {
			return nil, err
		}
		bids = append(bids, bid)
	}
	return bids, nil
}

func (r *BidRepository) CountByItemID(itemID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM bids WHERE item_id = $1`, itemID).Scan(&count)
	return count, err
}

func (r *BidRepository) UpdateStatus(id int64, status string) error {
	_, err := r.db.Exec(`UPDATE bids SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (r *BidRepository) MarkItemBidsOutbid(itemID int64, exceptBidID int64) error {
	query := `UPDATE bids SET status = 'outbid' WHERE item_id = $1 AND id != $2 AND status = 'active'`
	_, err := r.db.Exec(query, itemID, exceptBidID)
	return err
}
