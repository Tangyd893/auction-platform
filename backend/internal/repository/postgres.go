package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"auction-platform/internal/config"
)

func NewPostgresDB(cfg *config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'buyer',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS items (
			id SERIAL PRIMARY KEY,
			title VARCHAR(200) NOT NULL,
			description TEXT,
			image_url VARCHAR(500),
			start_price BIGINT NOT NULL,
			current_price BIGINT NOT NULL,
			reserve_price BIGINT NOT NULL DEFAULT 0,
			bid_increment BIGINT NOT NULL DEFAULT 100,
			seller_id BIGINT NOT NULL REFERENCES users(id),
			status VARCHAR(20) NOT NULL DEFAULT 'draft',
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS bids (
			id SERIAL PRIMARY KEY,
			item_id BIGINT NOT NULL REFERENCES items(id),
			buyer_id BIGINT NOT NULL REFERENCES users(id),
			amount BIGINT NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			item_id BIGINT NOT NULL REFERENCES items(id),
			seller_id BIGINT NOT NULL REFERENCES users(id),
			buyer_id BIGINT NOT NULL REFERENCES users(id),
			final_price BIGINT NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			paid_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_items_seller_id ON items(seller_id)`,
		`CREATE INDEX IF NOT EXISTS idx_items_status ON items(status)`,
		`CREATE INDEX IF NOT EXISTS idx_bids_item_id ON bids(item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bids_buyer_id ON bids(buyer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_buyer_id ON orders(buyer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_seller_id ON orders(seller_id)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
