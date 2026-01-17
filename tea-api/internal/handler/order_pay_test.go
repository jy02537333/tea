package handler

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

func TestOrderPay_MarksOrderPaid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	origDB := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = origDB
	})

	if err := db.AutoMigrate(&model.User{}, &model.Order{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	user := &model.User{OpenID: "pay-user", Phone: "13800000001"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	order := &model.Order{
		OrderNo:      "TPAY0001",
		UserID:       user.ID,
		StoreID:      1,
		TableID:      1,
		TableNo:      "A12",
		TotalAmount:  decimal.NewFromFloat(10),
		PayAmount:    decimal.NewFromFloat(10),
		Status:       1,
		PayStatus:    1,
		OrderType:    2,
		DeliveryType: 1,
		AddressInfo:  "{}",
	}
	if err := db.Create(order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	h := NewOrderHandler()
	r := gin.New()
	r.POST("/api/v1/orders/:id/pay", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		h.Pay(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/"+strconv.FormatUint(uint64(order.ID), 10)+"/pay", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var got model.Order
	if err := db.First(&got, order.ID).Error; err != nil {
		t.Fatalf("reload order: %v", err)
	}
	if got.Status != 2 || got.PayStatus != 2 {
		t.Fatalf("expected status=2 pay_status=2, got status=%d pay_status=%d", got.Status, got.PayStatus)
	}
	if got.PaidAt == nil {
		t.Fatalf("expected paid_at to be set")
	}
	// ensure table info preserved
	if got.TableID != 1 || got.TableNo != "A12" {
		t.Fatalf("expected table_id=1 table_no=A12, got table_id=%d table_no=%q", got.TableID, got.TableNo)
	}
}

func TestOrderPay_RejectsOtherUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	origDB := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = origDB
	})

	if err := db.AutoMigrate(&model.User{}, &model.Order{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	owner := &model.User{OpenID: "pay-owner", Phone: "13800000002"}
	if err := db.Create(owner).Error; err != nil {
		t.Fatalf("create owner: %v", err)
	}

	other := &model.User{OpenID: "pay-other", Phone: "13800000003"}
	if err := db.Create(other).Error; err != nil {
		t.Fatalf("create other: %v", err)
	}

	order := &model.Order{
		OrderNo:      "TPAY0002",
		UserID:       owner.ID,
		StoreID:      1,
		TotalAmount:  decimal.NewFromFloat(10),
		PayAmount:    decimal.NewFromFloat(10),
		Status:       1,
		PayStatus:    1,
		OrderType:    2,
		DeliveryType: 1,
		AddressInfo:  "{}",
	}
	if err := db.Create(order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	h := NewOrderHandler()
	r := gin.New()
	r.POST("/api/v1/orders/:id/pay", func(c *gin.Context) {
		c.Set("user_id", other.ID)
		h.Pay(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/"+strconv.FormatUint(uint64(order.ID), 10)+"/pay", nil)
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 for other user, got %d, body=%s", w.Code, w.Body.String())
	}

	var got model.Order
	if err := db.First(&got, order.ID).Error; err != nil {
		t.Fatalf("reload order: %v", err)
	}
	if got.Status != 1 || got.PayStatus != 1 {
		t.Fatalf("order should remain unpaid, got status=%d pay_status=%d", got.Status, got.PayStatus)
	}
	if got.PaidAt != nil {
		t.Fatalf("paid_at should remain nil")
	}
}
