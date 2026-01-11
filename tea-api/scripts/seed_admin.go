//go:build tools
// +build tools

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
	flag.StringVar(&cfgPath, "config", "configs/config.yaml", "path to config.yaml (relative to tea-api)")
	flag.Parse()

	if err := config.LoadConfig(cfgPath); err != nil {
		log.Fatalf("load config: %v", err)
	}

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
		password = "Admin@123"
	}

	phone := "13800000001"

	var existing model.User
	if err := database.GetDB().Where("username = ?", username).First(&existing).Error; err == nil && existing.ID != 0 {
		fmt.Printf("admin user already exists: id=%d username=%s\n", existing.ID, existing.Username)
		return
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	user := model.User{
		Username:     &username,
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
