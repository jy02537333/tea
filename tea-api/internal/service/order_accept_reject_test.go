package service

import (
    "fmt"
    "testing"
    "time"

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
    if err := db.AutoMigrate(&model.Order{}, &model.OrderItem{}, &model.UserCoupon{}, &model.Coupon{}); err != nil {
        t.Fatalf("auto migrate: %v", err)
    }
    database.DB = db
    return db
}

func uniqOrderNo() string { return fmt.Sprintf("UT-%d", time.Now().UnixNano()) }

func TestAdminAcceptOrder_Status2To3(t *testing.T) {
    db := setupOrderTestDB(t)

    // create paid order
    o := model.Order{OrderNo: uniqOrderNo(), UserID: 1, StoreID: 0, Status: 2, PayStatus: 2, OrderType: 1, DeliveryType: 1}
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

    // 规则更新：仅当状态=3(处理中/配送中) 才允许拒绝
    o := model.Order{OrderNo: uniqOrderNo(), UserID: 1, StoreID: 0, Status: 3, PayStatus: 2, OrderType: 1, DeliveryType: 1}
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

func TestAdminAcceptOrder_InvalidStatuses_ShouldError(t *testing.T) {
    db := setupOrderTestDB(t)

    cases := []int{1, 3, 4, 5}
    for _, st := range cases {
        o := model.Order{OrderNo: uniqOrderNo(), UserID: 1, StoreID: 0, Status: st, PayStatus: 2, OrderType: 1, DeliveryType: 1}
        if err := db.Create(&o).Error; err != nil {
            t.Fatalf("create order: %v", err)
        }
        svc := NewOrderService()
        if err := svc.AdminAcceptOrder(o.ID); err == nil {
            t.Fatalf("AdminAcceptOrder should error for status=%d", st)
        }
    }
}

func TestAdminRejectOrder_OnlyStatus3Allowed(t *testing.T) {
    db := setupOrderTestDB(t)

    // 非法状态：1、2、4、5 应报错
    badStatuses := []int{1, 2, 4, 5}
    for _, st := range badStatuses {
        o := model.Order{OrderNo: uniqOrderNo(), UserID: 1, StoreID: 0, Status: st, PayStatus: 2, OrderType: 1, DeliveryType: 1}
        if err := db.Create(&o).Error; err != nil {
            t.Fatalf("create order: %v", err)
        }
        svc := NewOrderService()
        if err := svc.AdminRejectOrder(o.ID, "reason"); err == nil {
            t.Fatalf("AdminRejectOrder should error for status=%d", st)
        }
    }

    // 合法状态：3 应成功
    ok := model.Order{OrderNo: uniqOrderNo(), UserID: 1, StoreID: 0, Status: 3, PayStatus: 2, OrderType: 1, DeliveryType: 1}
    if err := db.Create(&ok).Error; err != nil {
        t.Fatalf("create order ok: %v", err)
    }
    svc := NewOrderService()
    if err := svc.AdminRejectOrder(ok.ID, "reason"); err != nil {
        t.Fatalf("AdminRejectOrder should succeed for status=3, err=%v", err)
    }
}
