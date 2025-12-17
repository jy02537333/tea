package main

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

// grant a threshold coupon to user_id=1 for testing available-coupons
func main() {
	db := database.GetDB()

	c := &model.Coupon{
		Name:        "测试满减50-门槛300",
		Type:        1, // 满减券
		Amount:      decimal.NewFromInt(50),
		MinAmount:   decimal.NewFromInt(300),
		TotalCount:  100,
		Status:      1,
		StartTime:   time.Now().Add(-time.Hour),
		EndTime:     time.Now().Add(7 * 24 * time.Hour),
		Description: "自动化测试券",
	}
	if err := db.Create(c).Error; err != nil {
		panic(err)
	}
	uc := &model.UserCoupon{UserID: 1, CouponID: c.ID, Status: 1}
	if err := db.Create(uc).Error; err != nil {
		panic(err)
	}
	fmt.Printf("created coupon_id=%d user_coupon_id=%d\n", c.ID, uc.ID)
}
