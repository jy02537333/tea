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

func Test_Order_List_Filter_By_Store(t *testing.T) {
	// sqlite env vars removed; tests use MySQL by default
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 登录
	loginReq := map[string]string{"openid": "user_openid_list_filter"}
	b, _ := json.Marshal(loginReq)
	resp, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("dev-login request err: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("dev-login status: %d", resp.StatusCode)
	}
	var login struct {
		Code int
		Data struct{ Token string }
	}
	json.NewDecoder(resp.Body).Decode(&login)
	resp.Body.Close()
	if login.Code != 0 || login.Data.Token == "" {
		t.Fatalf("login failed: %+v", login)
	}
	auth := "Bearer " + login.Data.Token
	clearCart := func() {
		req, _ := http.NewRequest("DELETE", ts.URL+"/api/v1/cart/clear", nil)
		req.Header.Set("Authorization", auth)
		rc, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("clear cart err: %v", err)
		}
		if rc.StatusCode != 200 {
			t.Fatalf("clear cart status: %d", rc.StatusCode)
		}
		rc.Body.Close()
	}
	clearCart()

	// 获取管理员 token 以便绑定门店商品库存
	adminReq := map[string]string{"openid": "admin_openid"}
	adminBody, _ := json.Marshal(adminReq)
	adminResp, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(adminBody))
	if err != nil {
		t.Fatalf("admin dev-login request err: %v", err)
	}
	if adminResp.StatusCode != 200 {
		t.Fatalf("admin dev-login status: %d", adminResp.StatusCode)
	}
	var adminLogin struct {
		Code int
		Data struct{ Token string }
	}
	json.NewDecoder(adminResp.Body).Decode(&adminLogin)
	adminResp.Body.Close()
	if adminLogin.Code != 0 || adminLogin.Data.Token == "" {
		t.Fatalf("admin login failed: %+v", adminLogin)
	}
	adminAuth := "Bearer " + adminLogin.Data.Token

	// 创建两个门店
	createStore := func(name string) uint {
		st := map[string]any{"name": name, "address": "A", "latitude": 31.2, "longitude": 121.4, "status": 1}
		sb, _ := json.Marshal(st)
		req, _ := http.NewRequest("POST", ts.URL+"/api/v1/stores", bytes.NewReader(sb))
		req.Header.Set("Authorization", auth)
		req.Header.Set("Content-Type", "application/json")
		rs, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("create store err: %v", err)
		}
		if rs.StatusCode != 200 {
			t.Fatalf("create store status: %d", rs.StatusCode)
		}
		var storeResp struct {
			Code int
			Data struct{ ID uint }
		}
		json.NewDecoder(rs.Body).Decode(&storeResp)
		rs.Body.Close()
		return storeResp.Data.ID
	}
	s1 := createStore("门店1")
	s2 := createStore("门店2")

	// 创建分类与商品
	catReq := map[string]any{"name": "筛选分类"}
	cb, _ := json.Marshal(catReq)
	req, _ := http.NewRequest("POST", ts.URL+"/api/v1/categories", bytes.NewReader(cb))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create category err: %v", err)
	}
	if resp2.StatusCode != 200 {
		t.Fatalf("create category status: %d", resp2.StatusCode)
	}
	var catResp struct {
		Code int
		Data struct{ ID uint }
	}
	json.NewDecoder(resp2.Body).Decode(&catResp)
	resp2.Body.Close()

	prodReq := map[string]any{"category_id": catResp.Data.ID, "name": "筛选商品", "price": 10.00, "stock": 10, "status": 1}
	pb, _ := json.Marshal(prodReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/products", bytes.NewReader(pb))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	resp3, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create product err: %v", err)
	}
	if resp3.StatusCode != 200 {
		t.Fatalf("create product status: %d", resp3.StatusCode)
	}
	var prodResp struct {
		Code int
		Data struct{ ID uint }
	}
	json.NewDecoder(resp3.Body).Decode(&prodResp)
	resp3.Body.Close()

	bindProductToStore := func(storeID uint) {
		payload := map[string]any{
			"product_id":     prodResp.Data.ID,
			"stock":          5,
			"price_override": "10",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/admin/stores/%d/products", ts.URL, storeID), bytes.NewReader(body))
		req.Header.Set("Authorization", adminAuth)
		req.Header.Set("Content-Type", "application/json")
		rsp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("assign product to store %d err: %v", storeID, err)
		}
		if rsp.StatusCode != 200 {
			t.Fatalf("assign product to store %d status: %d", storeID, rsp.StatusCode)
		}
		rsp.Body.Close()
	}
	bindProductToStore(s1)
	bindProductToStore(s2)

	// 下单到门店1
	add := map[string]any{"product_id": prodResp.Data.ID, "quantity": 1}
	ab, _ := json.Marshal(add)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	if _, err := http.DefaultClient.Do(req); err != nil {
		t.Fatalf("add cart err: %v", err)
	}

	orderReq := map[string]any{"delivery_type": 1, "store_id": s1}
	ob, _ := json.Marshal(orderReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/orders/from-cart", bytes.NewReader(ob))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	respO1, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create order s1 err: %v", err)
	}
	var o1 struct {
		Code int
		Data struct{ ID uint }
	}
	json.NewDecoder(respO1.Body).Decode(&o1)
	respO1.Body.Close()
	if o1.Code != 0 || o1.Data.ID == 0 {
		t.Fatalf("create order s1 failed")
	}

	// 下单到门店2
	ab2, _ := json.Marshal(add)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab2))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	if _, err := http.DefaultClient.Do(req); err != nil {
		t.Fatalf("add cart 2 err: %v", err)
	}
	orderReq2 := map[string]any{"delivery_type": 1, "store_id": s2}
	ob2, _ := json.Marshal(orderReq2)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/orders/from-cart", bytes.NewReader(ob2))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	respO2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create order s2 err: %v", err)
	}
	var o2 struct {
		Code int
		Data struct{ ID uint }
	}
	json.NewDecoder(respO2.Body).Decode(&o2)
	respO2.Body.Close()
	if o2.Code != 0 || o2.Data.ID == 0 {
		t.Fatalf("create order s2 failed")
	}

	// 按门店1筛选订单
	url := fmt.Sprintf("%s/api/v1/orders?store_id=%d", ts.URL, s1)
	req, _ = http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", auth)
	respL, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("list orders err: %v", err)
	}
	if respL.StatusCode != 200 {
		t.Fatalf("list orders status: %d", respL.StatusCode)
	}
	var listResp struct {
		Code int
		Data []struct {
			ID      uint
			StoreID uint `json:"store_id"`
		}
		Total int64
	}
	json.NewDecoder(respL.Body).Decode(&listResp)
	respL.Body.Close()
	if listResp.Code != 0 {
		t.Fatalf("list orders code=%d", listResp.Code)
	}
	if len(listResp.Data) != 1 {
		t.Fatalf("expect 1 order for store1, got %d", len(listResp.Data))
	}
	if listResp.Data[0].StoreID != s1 {
		t.Fatalf("store_id mismatch: %d", listResp.Data[0].StoreID)
	}
}
