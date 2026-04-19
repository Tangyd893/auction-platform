package main

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 创建默认管理员用户
	dsn := "host=localhost port=5432 user=postgres password=postgres123 dbname=auction sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		os.Exit(1)
	}
	defer db.Close()

	// 检查是否已存在管理员
	var exists bool
	err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", "admin")
	if err != nil {
		fmt.Println("Failed to check admin:", err)
		os.Exit(1)
	}

	if exists {
		fmt.Println("Admin user already exists")
		return
	}

	// 创建管理员
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	_, err = db.Exec(`INSERT INTO users (username, password, email, role) VALUES ($1, $2, $3, $4)`,
		"admin", string(hashedPassword), "admin@auction.com", "admin")
	if err != nil {
		fmt.Println("Failed to create admin:", err)
		os.Exit(1)
	}

	fmt.Println("Admin user created successfully")
	fmt.Println("Username: admin")
	fmt.Println("Password: admin123")
}
