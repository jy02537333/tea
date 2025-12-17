package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
	"time"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "configs/config.yaml", "path to tea-api config.yaml")
	flag.Parse()

	// If default path not found and looks relative, try repo-root fallback
	if _, err := os.Stat(cfgPath); err != nil {
		alt := filepath.Join("..", "tea-api", "configs", "config.yaml")
		if _, err2 := os.Stat(alt); err2 == nil {
			cfgPath = alt
		}
	}

	if err := config.LoadConfig(cfgPath); err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Connect to DB without running migrations
	db, err := database.InitWithoutMigrate()
	if err != nil {
		log.Fatalf("init db: %v", err)
	}
	_ = db

	username := os.Getenv("SEED_ADMIN_USER")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("SEED_ADMIN_PASS")
	if password == "" {
		password = "Admin@123" // change after first login
	}

	// Prefer an explicit phone if provided; otherwise generate a numeric, login-friendly phone
	phone := os.Getenv("SEED_ADMIN_PHONE")
	if phone == "" {
		// Generate a pseudo phone: 139 + 8 digits from timestamp
		n := time.Now().UnixNano() % 100000000
		phone = fmt.Sprintf("139%08d", n)
	}

	// Check if user exists
	var existing model.User
	if err := database.GetDB().Where("username = ? OR phone = ?", username, phone).First(&existing).Error; err == nil && existing.ID != 0 {
		fmt.Printf("admin user already exists: id=%d username=%s\n", existing.ID, existing.Username)
		return
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	user := model.User{
		Username:     username,
		PasswordHash: hash,
		Phone:        phone,
		Nickname:     "管理员",
		Role:         "admin",
		Status:       1,
	}

	if err := database.GetDB().Create(&user).Error; err != nil {
		log.Fatalf("create admin user: %v", err)
	}

	fmt.Printf("created admin user id=%d username=%s phone=%s role=%s\n", user.ID, user.Username, user.Phone, user.Role)
}
