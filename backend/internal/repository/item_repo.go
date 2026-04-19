package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"auction-platform/internal/model"
)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) Create(item *model.Item) error {
	query := `INSERT INTO items (title, description, image_url, start_price, current_price, reserve_price, bid_increment, seller_id, status, start_time, end_time)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, item.Title, item.Description, item.ImageURL, item.StartPrice, item.CurrentPrice,
		item.ReservePrice, item.BidIncrement, item.SellerID, item.Status, item.StartTime, item.EndTime).
		Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
}

func (r *ItemRepository) GetByID(id int64) (*model.Item, error) {
	query := `SELECT id, title, description, image_url, start_price, current_price, reserve_price, bid_increment, seller_id, status, start_time, end_time, created_at, updated_at
			  FROM items WHERE id = $1`
	item := &model.Item{}
	err := r.db.QueryRow(query, id).Scan(&item.ID, &item.Title, &item.Description, &item.ImageURL,
		&item.StartPrice, &item.CurrentPrice, &item.ReservePrice, &item.BidIncrement, &item.SellerID,
		&item.Status, &item.StartTime, &item.EndTime, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *ItemRepository) List(status string, sellerID int64, keyword string, page, pageSize int) ([]*model.Item, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if sellerID > 0 {
		conditions = append(conditions, fmt.Sprintf("seller_id = $%d", argIdx))
		args = append(args, sellerID)
		argIdx++
	}
	if keyword != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+keyword+"%")
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM items %s`, whereClause)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT id, title, description, image_url, start_price, current_price, reserve_price, bid_increment, seller_id, status, start_time, end_time, created_at, updated_at
			  FROM items %s ORDER BY id DESC LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.Item
	for rows.Next() {
		item := &model.Item{}
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.ImageURL,
			&item.StartPrice, &item.CurrentPrice, &item.ReservePrice, &item.BidIncrement, &item.SellerID,
			&item.Status, &item.StartTime, &item.EndTime, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, nil
}

func (r *ItemRepository) Update(item *model.Item) error {
	query := `UPDATE items SET title = $1, description = $2, image_url = $3, reserve_price = $4, bid_increment = $5, end_time = $6, status = $7, updated_at = CURRENT_TIMESTAMP WHERE id = $8`
	_, err := r.db.Exec(query, item.Title, item.Description, item.ImageURL, item.ReservePrice, item.BidIncrement, item.EndTime, item.Status, item.ID)
	return err
}

func (r *ItemRepository) UpdatePrice(id, price int64) error {
	query := `UPDATE items SET current_price = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.Exec(query, price, id)
	return err
}

func (r *ItemRepository) UpdateStatus(id int64, status string) error {
	query := `UPDATE items SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *ItemRepository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM items WHERE id = $1`, id)
	return err
}
