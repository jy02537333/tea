package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "tea-api/configs/config.yaml", "path to tea-api config.yaml")
	flag.Parse()

	// Allow override via TEA_DSN and other TEA_ env vars
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

	// Avoid duplicate phone; generate a unique phone-like value
	phone := "seed-" + utils.GenerateUID()

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

	fmt.Printf("created admin user id=%d username=%s phone=%s\n", user.ID, user.Username, user.Phone)
}
