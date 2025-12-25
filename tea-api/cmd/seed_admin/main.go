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
	"time"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "configs/config.yaml", "path to tea-api config.yaml")
	flag.Parse()

	if err := config.LoadConfig(cfgPath); err != nil {
		log.Fatalf("load config: %v", err)
	}

	if _, err := database.InitWithoutMigrate(); err != nil {
		log.Fatalf("init db: %v", err)
	}

	username := os.Getenv("SEED_ADMIN_USER")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("SEED_ADMIN_PASS")
	if password == "" {
		password = "Admin@123"
	}

	phone := os.Getenv("SEED_ADMIN_PHONE")
	if phone == "" {
		n := time.Now().UnixNano() % 100000000
		phone = fmt.Sprintf("139%08d", n)
	}

	var existing model.User
	if err := database.GetDB().Where("username = ? OR phone = ?", username, phone).First(&existing).Error; err == nil && existing.ID != 0 {
		fmt.Printf("admin user already exists: id=%d username=%s phone=%s role=%s\n", existing.ID, derefString(existing.Username), existing.Phone, existing.Role)
		return
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	uname := username
	user := model.User{
		Username:     &uname,
		PasswordHash: hash,
		OpenID:       "seed-admin-" + utils.GenerateUID(),
		Phone:        phone,
		Nickname:     "管理员",
		Role:         "admin",
		Status:       1,
	}
	if err := database.GetDB().Create(&user).Error; err != nil {
		log.Fatalf("create admin user: %v", err)
	}
	fmt.Printf("created admin user id=%d username=%s phone=%s role=%s\n", user.ID, derefString(user.Username), user.Phone, user.Role)
}

func derefString(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}
