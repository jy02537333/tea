package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"

	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/pkg/database"
)

// create a store-bound coupon (store_id=1) and grant to user_id=1
func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "configs/config.yaml", "path to tea-api config.yaml")
	flag.Parse()

	if err := config.LoadConfig(cfgPath); err != nil {
		log.Fatalf("load config: %v", err)
	}
	db, err := database.InitWithoutMigrate()
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	storeID := uint(1)
	c := &model.Coupon{
		Name:        "门店券-满减60-门槛300(店1)",
		Type:        1,
		Amount:      decimal.NewFromInt(60),
		MinAmount:   decimal.NewFromInt(300),
		TotalCount:  100,
		Status:      1,
		StartTime:   time.Now().Add(-time.Hour),
		EndTime:     time.Now().Add(7 * 24 * time.Hour),
		Description: "自动化门店测试券",
		StoreID:     &storeID,
	}
	if err := db.Create(c).Error; err != nil {
		panic(err)
	}
	uc := &model.UserCoupon{UserID: 1, CouponID: c.ID, Status: 1}
	if err := db.Create(uc).Error; err != nil {
		panic(err)
	}
	fmt.Printf("created store_coupon_id=%d user_coupon_id=%d store_id=%d\n", c.ID, uc.ID, storeID)
}
