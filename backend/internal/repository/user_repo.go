package repository

import (
	"database/sql"

	"auction-platform/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	query := `INSERT INTO users (username, password, email, role) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	return r.db.QueryRow(query, user.Username, user.Password, user.Email, user.Role).Scan(&user.ID, &user.CreatedAt)
}

func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	query := `SELECT id, username, password, email, role, created_at FROM users WHERE id = $1`
	user := &model.User{}
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	query := `SELECT id, username, password, email, role, created_at FROM users WHERE username = $1`
	user := &model.User{}
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) List(page, pageSize int) ([]*model.User, int, error) {
	offset := (page - 1) * pageSize
	query := `SELECT id, username, password, email, role, created_at FROM users ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	var total int
	r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&total)

	return users, total, nil
}

func (r *UserRepository) Update(user *model.User) error {
	query := `UPDATE users SET username = $1, email = $2, role = $3 WHERE id = $4`
	_, err := r.db.Exec(query, user.Username, user.Email, user.Role, user.ID)
	return err
}

func (r *UserRepository) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, username).Scan(&exists)
	return exists, err
}

func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}

func (r *UserRepository) InitDefaultAdmin() error {
	exists, err := r.ExistsByUsername("admin")
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// 密码: admin123 (bcrypt hash)
	admin := &model.User{
		Username: "admin",
		Password: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // admin123
		Email:    "admin@auction.com",
		Role:     "admin",
	}
	return r.Create(admin)
}
