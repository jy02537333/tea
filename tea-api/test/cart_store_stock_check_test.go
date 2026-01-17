package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"tea-api/internal/config"
	"tea-api/internal/router"
	"tea-api/pkg/database"
)

func Test_Cart_CheckStock_UsesStoreStock_WhenStoreIDProvided(t *testing.T) {
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// dev-login 普通用户
	b, _ := json.Marshal(map[string]string{"openid": "user_openid_cart_store_stock"})
	resp, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("dev-login request err: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("dev-login status: %d", resp.StatusCode)
	}
	var login struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&login); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if login.Code != 0 || login.Data.Token == "" {
		t.Fatalf("dev-login failed: %+v", login)
	}
	authHeader := "Bearer " + login.Data.Token

	// clear cart
	req, _ := http.NewRequest("DELETE", ts.URL+"/api/v1/cart/clear", nil)
	req.Header.Set("Authorization", authHeader)
	if respClr, err := http.DefaultClient.Do(req); err == nil {
		respClr.Body.Close()
	} else {
		t.Fatalf("clear cart err: %v", err)
	}

	// create store
	sb, _ := json.Marshal(map[string]any{"name": "门店-加购库存校验", "status": 1})
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/stores", bytes.NewReader(sb))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	respS, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create store err: %v", err)
	}
	var storeResp struct {
		Code int
		Data struct{ ID uint }
	}
	_ = json.NewDecoder(respS.Body).Decode(&storeResp)
	respS.Body.Close()
	if storeResp.Code != 0 || storeResp.Data.ID == 0 {
		t.Fatalf("invalid store resp: %+v", storeResp)
	}
	storeID := storeResp.Data.ID

	// create and bind a store-admin user for the store exclusive endpoints
	adminBody, _ := json.Marshal(map[string]string{"openid": "store_admin_openid_cart_store_stock"})
	respA, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(adminBody))
	if err != nil {
		t.Fatalf("dev-login store-admin err: %v", err)
	}
	var adminLogin struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	_ = json.NewDecoder(respA.Body).Decode(&adminLogin)
	respA.Body.Close()
	if adminLogin.Code != 0 || adminLogin.Data.Token == "" {
		t.Fatalf("dev-login store-admin failed: %+v", adminLogin)
	}
	adminAuth := "Bearer " + adminLogin.Data.Token
	// bind store_admins
	db := database.GetDB()
	if db == nil {
		t.Fatalf("db is nil")
	}
	var userRow struct {
		ID uint `gorm:"column:id"`
	}
	if err := db.Table("users").Select("id").Where("open_id = ?", "store_admin_openid_cart_store_stock").Order("id desc").Limit(1).Find(&userRow).Error; err != nil {
		t.Fatalf("query store-admin user err: %v", err)
	}
	if userRow.ID == 0 {
		t.Fatalf("store-admin user id is 0")
	}
	_ = db.Table("store_admins").Create(map[string]any{"user_id": userRow.ID, "store_id": storeID}).Error

	// create category
	cb, _ := json.Marshal(map[string]any{"name": "门店库存分类"})
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/categories", bytes.NewReader(cb))
	req.Header.Set("Authorization", authHeader)
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

	// create store exclusive product (platform stock=0, store stock=2)
	exb, _ := json.Marshal(map[string]any{
		"name":        "门店特供-库存2",
		"category_id": catResp.Data.ID,
		"price":       "9.90",
		"stock":       2,
	})
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/stores/%d/exclusive-products/new", ts.URL, storeID), bytes.NewReader(exb))
	req.Header.Set("Authorization", adminAuth)
	req.Header.Set("Content-Type", "application/json")
	respX, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create exclusive product err: %v", err)
	}
	var exResp struct {
		Code int
		Data struct {
			ProductID uint `json:"product_id"`
		} `json:"data"`
		Message string `json:"message"`
	}
	_ = json.NewDecoder(respX.Body).Decode(&exResp)
	respX.Body.Close()
	if exResp.Code != 0 || exResp.Data.ProductID == 0 {
		t.Fatalf("invalid exclusive product resp: %+v", exResp)
	}
	pid := exResp.Data.ProductID

	// store context: quantity > store stock should be blocked
	abTooMuch, _ := json.Marshal(map[string]any{"product_id": pid, "quantity": 3})
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/cart/items?store_id=%d", ts.URL, storeID), bytes.NewReader(abTooMuch))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	respBad, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("add cart too much err: %v", err)
	}
	var badResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	_ = json.NewDecoder(respBad.Body).Decode(&badResp)
	respBad.Body.Close()
	if respBad.StatusCode == 200 && badResp.Code == 0 {
		t.Fatalf("expected failure when quantity exceeds store stock")
	}
	if badResp.Message == "" {
		t.Fatalf("expected error message when quantity exceeds store stock")
	}

	// store context: within store stock should pass
	abOK, _ := json.Marshal(map[string]any{"product_id": pid, "quantity": 1})
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/cart/items?store_id=%d", ts.URL, storeID), bytes.NewReader(abOK))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	respOK, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("add cart ok err: %v", err)
	}
	var okResp struct {
		Code int `json:"code"`
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
		Message string `json:"message"`
	}
	_ = json.NewDecoder(respOK.Body).Decode(&okResp)
	respOK.Body.Close()
	if respOK.StatusCode != 200 || okResp.Code != 0 || okResp.Data.ID == 0 {
		t.Fatalf("expected add cart success, status=%d resp=%+v", respOK.StatusCode, okResp)
	}

	// platform context (no store_id): should fail because product.stock is 0 for store-exclusive product
	abNoStore, _ := json.Marshal(map[string]any{"product_id": pid, "quantity": 1})
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(abNoStore))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	respNoStore, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("add cart no store err: %v", err)
	}
	var nsResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	_ = json.NewDecoder(respNoStore.Body).Decode(&nsResp)
	respNoStore.Body.Close()
	if respNoStore.StatusCode == 200 && nsResp.Code == 0 {
		t.Fatalf("expected failure when adding store-exclusive product without store_id")
	}
}
