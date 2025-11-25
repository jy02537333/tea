package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/shopspring/decimal"

	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

func main() {
	openid := flag.String("openid", "", "user openid")
	phone := flag.String("phone", "", "user phone")
	role := flag.String("role", "user", "user role (user/admin)")
	flag.Parse()

	if *openid == "" {
		fmt.Println("-openid required")
		os.Exit(2)
	}
	if err := config.LoadConfig("configs/config.yaml"); err != nil {
		fmt.Printf("load config: %v\n", err)
		os.Exit(2)
	}
	database.InitDatabase()
	db := database.GetDB()

	p := *phone
	if p == "" {
		p = fmt.Sprintf("18%09d", time.Now().UnixNano()%1_000_000_000)
	}
	u := &model.User{
		BaseModel:    model.BaseModel{UID: utils.GenerateUID()},
		OpenID:       *openid,
		Phone:        p,
		Nickname:     "E2E User",
		Status:       1,
		Balance:      decimal.NewFromInt(0),
		InterestRate: decimal.NewFromInt(0),
		Points:       0,
		Role:         *role,
	}
	if err := db.Where("open_id = ?", u.OpenID).FirstOrCreate(u).Error; err != nil {
		fmt.Printf("create user err: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("ok: user id=%d openid=%s phone=%s role=%s\n", u.ID, u.OpenID, u.Phone, u.Role)
}
