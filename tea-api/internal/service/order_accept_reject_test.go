package service

import (
    "testing"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"

    "tea-api/internal/model"
    "tea-api/pkg/database"
)

func setupOrderTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    db, err := gorm.Open(sqlite.Open("file:order_test?mode=memory&cache=shared"), &gorm.Config{})
    if err != nil {
        t.Fatalf("open sqlite: %v", err)
    }
    // migrate minimal tables required by AdminAccept/Reject
    if err := db.AutoMigrate(&model.Order{}, &model.UserCoupon{}, &model.Coupon{}); err != nil {
        t.Fatalf("auto migrate: %v", err)
    }
    database.DB = db
    return db
}

func TestAdminAcceptOrder_Status2To3(t *testing.T) {
    db := setupOrderTestDB(t)

    // create paid order
    o := model.Order{UserID: 1, StoreID: 0, Status: 2, PayStatus: 2, OrderType: 1, DeliveryType: 1}
    if err := db.Create(&o).Error; err != nil {
        t.Fatalf("create order: %v", err)
    }

    svc := NewOrderService()
    if err := svc.AdminAcceptOrder(o.ID); err != nil {
        t.Fatalf("AdminAcceptOrder err: %v", err)
    }
    var got model.Order
    if err := db.First(&got, o.ID).Error; err != nil {
        t.Fatalf("query order: %v", err)
    }
    if got.Status != 3 {
        t.Fatalf("status not moved to 3, got=%d", got.Status)
    }
    if got.PayStatus != 2 {
        t.Fatalf("pay_status changed unexpectedly, got=%d", got.PayStatus)
    }
}

func TestAdminRejectOrder_StatusToCancelledRefunded(t *testing.T) {
    db := setupOrderTestDB(t)

    o := model.Order{UserID: 1, StoreID: 0, Status: 2, PayStatus: 2, OrderType: 1, DeliveryType: 1}
    if err := db.Create(&o).Error; err != nil {
        t.Fatalf("create order: %v", err)
    }

    svc := NewOrderService()
    if err := svc.AdminRejectOrder(o.ID, "no_stock"); err != nil {
        t.Fatalf("AdminRejectOrder err: %v", err)
    }
    var got model.Order
    if err := db.First(&got, o.ID).Error; err != nil {
        t.Fatalf("query order: %v", err)
    }
    if got.Status != 5 {
        t.Fatalf("status not set to cancelled(5), got=%d", got.Status)
    }
    if got.PayStatus != 4 {
        t.Fatalf("pay_status not set to refunded(4), got=%d", got.PayStatus)
    }
}
