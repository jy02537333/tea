package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

func TestAdminSetTable_UpdatesOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// Bind the in-memory DB to the global used by handlers.
	origDB := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = origDB
	})

	if err := db.AutoMigrate(&model.User{}, &model.Order{}, &model.OperationLog{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	user := &model.User{OpenID: "test-openid", Phone: "13800000000"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	order := &model.Order{
		OrderNo:       "TTEST0001",
		UserID:        user.ID,
		StoreID:       1,
		TableID:       0,
		TableNo:       "",
		TotalAmount:   decimal.NewFromFloat(10),
		PayAmount:     decimal.NewFromFloat(10),
		Status:        1,
		PayStatus:     1,
		OrderType:     2,
		DeliveryType:  1,
		AddressInfo:   "{}",
		Remark:        "",
		CancelReason:  "",
	}

	if err := db.Create(order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	h := NewOrderHandler()
	r := gin.New()
	r.POST("/api/v1/orders/:id/set-table", func(c *gin.Context) {
		c.Set("user_id", uint(999))
		h.AdminSetTable(c)
	})

	body := map[string]any{
		"table_id":  12,
		"table_no":  " A12 ",
		"reason":    "fix",
	}
	bs, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/"+itoa(order.ID)+"/set-table", bytes.NewReader(bs))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			OK bool `json:"ok"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v, body=%s", err, w.Body.String())
	}
	if resp.Code != 0 || !resp.Data.OK {
		t.Fatalf("unexpected response: %+v", resp)
	}

	var got model.Order
	if err := db.First(&got, order.ID).Error; err != nil {
		t.Fatalf("reload order: %v", err)
	}
	if got.TableID != 12 || got.TableNo != "A12" {
		t.Fatalf("expected table_id=12 table_no=A12, got table_id=%d table_no=%q", got.TableID, got.TableNo)
	}

	var opCount int64
	if err := db.Model(&model.OperationLog{}).Where("operation = ?", "order.set_table").Count(&opCount).Error; err != nil {
		t.Fatalf("count op logs: %v", err)
	}
	if opCount != 1 {
		t.Fatalf("expected 1 op log, got %d", opCount)
	}
}

func itoa(v uint) string {
	// tiny local helper to avoid strconv import in test file
	if v == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for v > 0 {
		d := v % 10
		buf = append(buf, byte('0'+d))
		v /= 10
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
