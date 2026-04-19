package repository

import (
	"database/sql"
	"fmt"

	"auction-platform/internal/model"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(order *model.Order) error {
	query := `INSERT INTO orders (item_id, seller_id, buyer_id, final_price, status) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.db.QueryRow(query, order.ItemID, order.SellerID, order.BuyerID, order.FinalPrice, order.Status).Scan(&order.ID, &order.CreatedAt)
}

func (r *OrderRepository) GetByID(id int64) (*model.Order, error) {
	query := `SELECT id, item_id, seller_id, buyer_id, final_price, status, created_at, paid_at FROM orders WHERE id = $1`
	order := &model.Order{}
	err := r.db.QueryRow(query, id).Scan(&order.ID, &order.ItemID, &order.SellerID, &order.BuyerID, &order.FinalPrice, &order.Status, &order.CreatedAt, &order.PaidAt)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *OrderRepository) List(userID int64, status string, page, pageSize int) ([]*model.Order, int, error) {
	conditions := []string{}
	args := []interface{}{}
	argIdx := 1

	if userID > 0 {
		conditions = append(conditions, fmt.Sprintf("(seller_id = $%d OR buyer_id = $%d)", argIdx, argIdx))
		args = append(args, userID)
		argIdx++
	}
	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM orders %s`, whereClause)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT id, item_id, seller_id, buyer_id, final_price, status, created_at, paid_at FROM orders %s ORDER BY id DESC LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		o := &model.Order{}
		if err := rows.Scan(&o.ID, &o.ItemID, &o.SellerID, &o.BuyerID, &o.FinalPrice, &o.Status, &o.CreatedAt, &o.PaidAt); err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}

	return orders, total, nil
}

func (r *OrderRepository) UpdateStatus(id int64, status string) error {
	query := `UPDATE orders SET status = $1`
	if status == "paid" {
		query += `, paid_at = CURRENT_TIMESTAMP`
	}
	query += ` WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *OrderRepository) GetByItemID(itemID int64) (*model.Order, error) {
	query := `SELECT id, item_id, seller_id, buyer_id, final_price, status, created_at, paid_at FROM orders WHERE item_id = $1`
	order := &model.Order{}
	err := r.db.QueryRow(query, itemID).Scan(&order.ID, &order.ItemID, &order.SellerID, &order.BuyerID, &order.FinalPrice, &order.Status, &order.CreatedAt, &order.PaidAt)
	if err != nil {
		return nil, err
	}
	return order, nil
}
