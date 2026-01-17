package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tea-api/internal/config"
	"tea-api/internal/router"
	"tea-api/internal/service"
	"tea-api/pkg/database"
)

func Test_Cart_Delete_RestoresStock_And_CleanupExpiredRestoresStock(t *testing.T) {
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// dev-login 普通用户
	b, _ := json.Marshal(map[string]string{"openid": "user_openid_cart_reserve_restore"})
	resp, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("dev-login err: %v", err)
	}
	defer resp.Body.Close()
	var login struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&login)
	if login.Code != 0 || login.Data.Token == "" {
		t.Fatalf("dev-login failed: %+v", login)
	}
	auth := "Bearer " + login.Data.Token

	// clear cart
	req, _ := http.NewRequest("DELETE", ts.URL+"/api/v1/cart/clear", nil)
	req.Header.Set("Authorization", auth)
	if respClr, err := http.DefaultClient.Do(req); err == nil {
		respClr.Body.Close()
	} else {
		t.Fatalf("clear cart err: %v", err)
	}

	// create category
	cb, _ := json.Marshal(map[string]any{"name": "预占库存分类"})
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/categories", bytes.NewReader(cb))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	respC, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create category err: %v", err)
	}
	var catResp struct {
		Code int
		Data struct{ ID uint }
	}
	_ = json.NewDecoder(respC.Body).Decode(&catResp)
	respC.Body.Close()
	if catResp.Code != 0 || catResp.Data.ID == 0 {
		t.Fatalf("invalid category resp: %+v", catResp)
	}

	// create product stock=5
	pb, _ := json.Marshal(map[string]any{"category_id": catResp.Data.ID, "name": "预占库存商品", "price": 9.9, "stock": 5, "status": 1})
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/products", bytes.NewReader(pb))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	respP, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create product err: %v", err)
	}
	var prodResp struct {
		Code int
		Data struct{ ID uint }
	}
	_ = json.NewDecoder(respP.Body).Decode(&prodResp)
	respP.Body.Close()
	if prodResp.Code != 0 || prodResp.Data.ID == 0 {
		t.Fatalf("invalid product resp: %+v", prodResp)
	}
	pid := prodResp.Data.ID

	getProductStock := func() int {
		db := database.GetDB()
		var row struct {
			Stock int `gorm:"column:stock"`
		}
		_ = db.Table("products").Select("stock").Where("id = ?", pid).Find(&row).Error
		return row.Stock
	}

	// add to cart quantity=2 (reserve)
	ab, _ := json.Marshal(map[string]any{"product_id": pid, "quantity": 2})
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	respAdd, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("add cart err: %v", err)
	}
	var addResp struct {
		Code int
		Data struct{ ID uint }
		Message string
	}
	_ = json.NewDecoder(respAdd.Body).Decode(&addResp)
	respAdd.Body.Close()
	if respAdd.StatusCode != 200 || addResp.Code != 0 || addResp.Data.ID == 0 {
		t.Fatalf("add cart failed: status=%d resp=%+v", respAdd.StatusCode, addResp)
	}
	if got := getProductStock(); got != 3 {
		t.Fatalf("expected stock reserved to 3, got %d", got)
	}

	// delete item -> restore to 5
	req, _ = http.NewRequest("DELETE", ts.URL+"/api/v1/cart/items/"+fmt.Sprintf("%d", addResp.Data.ID), nil)
	req.Header.Set("Authorization", auth)
	respDel, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("delete cart item err: %v", err)
	}
	respDel.Body.Close()
	if got := getProductStock(); got != 5 {
		t.Fatalf("expected stock restored to 5 after delete, got %d", got)
	}

	// add again then simulate expired + cleanup
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	respAdd2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("add cart 2 err: %v", err)
	}
	var add2 struct {
		Code int
		Data struct{ ID uint }
	}
	_ = json.NewDecoder(respAdd2.Body).Decode(&add2)
	respAdd2.Body.Close()
	if add2.Code != 0 || add2.Data.ID == 0 {
		t.Fatalf("add cart 2 failed: %+v", add2)
	}
	if got := getProductStock(); got != 3 {
		t.Fatalf("expected stock reserved to 3, got %d", got)
	}

	// backdate created_at to be expired
	db := database.GetDB()
	old := time.Now().Add(-31 * time.Minute)
	if err := db.Table("cart_items").Where("id = ?", add2.Data.ID).Update("created_at", old).Error; err != nil {
		t.Fatalf("update created_at err: %v", err)
	}

	svc := service.NewCartService()
	n, err := svc.CleanupExpiredItems(30*time.Minute, 100)
	if err != nil {
		t.Fatalf("cleanup err: %v", err)
	}
	if n <= 0 {
		t.Fatalf("expected cleanup deleted >=1 item")
	}
	if got := getProductStock(); got != 5 {
		t.Fatalf("expected stock restored to 5 after cleanup, got %d", got)
	}
}
