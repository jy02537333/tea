package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"tea-api/internal/config"
	"tea-api/internal/router"
	"tea-api/pkg/database"
)

func Test_Store_Order_Stats(t *testing.T) {
	_ = os.Setenv("TEA_USE_SQLITE", "0")
	_ = os.Setenv("TEA_SQLITE_PATH", "tea_test_store_stats.db")
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 普通用户登录
	loginReq := map[string]string{"openid": "user_openid_store_stats"}
	b, _ := json.Marshal(loginReq)
	resp, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("dev-login user err: %v", err)
	}
	var login struct {
		Code int
		Data struct{ Token string }
	}
	json.NewDecoder(resp.Body).Decode(&login)
	resp.Body.Close()
	if login.Data.Token == "" {
		t.Fatalf("user login failed")
	}
	userAuth := "Bearer " + login.Data.Token

	// 清理历史购物车，避免残留商品影响下单次数
	reqClr, _ := http.NewRequest("DELETE", ts.URL+"/api/v1/cart/clear", nil)
	reqClr.Header.Set("Authorization", userAuth)
	if respClr, err := http.DefaultClient.Do(reqClr); err != nil {
		t.Fatalf("clear cart err: %v", err)
	} else {
		respClr.Body.Close()
	}

	// 创建门店
	st := map[string]any{"name": "统计门店", "address": "X", "latitude": 31.2, "longitude": 121.4, "status": 1}
	sb, _ := json.Marshal(st)
	req, _ := http.NewRequest("POST", ts.URL+"/api/v1/stores", bytes.NewReader(sb))
	req.Header.Set("Authorization", userAuth)
	req.Header.Set("Content-Type", "application/json")
	respS, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create store err: %v", err)
	}
	var storeResp struct {
		Code int
		Data struct{ ID uint }
	}
	json.NewDecoder(respS.Body).Decode(&storeResp)
	respS.Body.Close()
	storeID := storeResp.Data.ID

	// 分类与商品
	catReq := map[string]any{"name": "统计分类"}
	cb, _ := json.Marshal(catReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/categories", bytes.NewReader(cb))
	req.Header.Set("Authorization", userAuth)
	req.Header.Set("Content-Type", "application/json")
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create category err: %v", err)
	}
	var catResp struct {
		Code int
		Data struct{ ID uint }
	}
	json.NewDecoder(resp2.Body).Decode(&catResp)
	resp2.Body.Close()

	prodReq := map[string]any{"category_id": catResp.Data.ID, "name": "统计商品", "price": 10.00, "stock": 10, "status": 1}
	pb, _ := json.Marshal(prodReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/products", bytes.NewReader(pb))
	req.Header.Set("Authorization", userAuth)
	req.Header.Set("Content-Type", "application/json")
	resp3, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create product err: %v", err)
	}
	var prodResp struct {
		Code int
		Data struct{ ID uint }
	}
	json.NewDecoder(resp3.Body).Decode(&prodResp)
	resp3.Body.Close()

	// 管理员登录并绑定门店商品库存
	adminReq := map[string]string{"openid": "admin_openid"}
	abA, _ := json.Marshal(adminReq)
	respA, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(abA))
	if err != nil {
		t.Fatalf("dev-login admin err: %v", err)
	}
	var adminLogin struct {
		Code int
		Data struct{ Token string }
	}
	json.NewDecoder(respA.Body).Decode(&adminLogin)
	respA.Body.Close()
	if adminLogin.Data.Token == "" {
		t.Fatalf("admin login failed")
	}
	adminAuth := "Bearer " + adminLogin.Data.Token

	upReq := map[string]any{"product_id": prodResp.Data.ID, "stock": 10, "price_override": "10.00"}
	upb, _ := json.Marshal(upReq)
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/admin/stores/%d/products", ts.URL, storeID), bytes.NewReader(upb))
	req.Header.Set("Authorization", adminAuth)
	req.Header.Set("Content-Type", "application/json")
	respUp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("bind store product err: %v", err)
	}
	respUp.Body.Close()

	// 下两单（各1件，单价10）
	add := map[string]any{"product_id": prodResp.Data.ID, "quantity": 1}
	ab, _ := json.Marshal(add)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab))
	req.Header.Set("Authorization", userAuth)
	req.Header.Set("Content-Type", "application/json")
	if _, err := http.DefaultClient.Do(req); err != nil {
		t.Fatalf("add cart err: %v", err)
	}
	orderReq := map[string]any{"delivery_type": 1, "store_id": storeID}
	ob, _ := json.Marshal(orderReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/orders/from-cart", bytes.NewReader(ob))
	req.Header.Set("Authorization", userAuth)
	req.Header.Set("Content-Type", "application/json")
	respO1, _ := http.DefaultClient.Do(req)
	var o1 struct {
		Code int
		Data struct{ ID uint }
	}
	json.NewDecoder(respO1.Body).Decode(&o1)
	respO1.Body.Close()
	if respO1.StatusCode != http.StatusOK || o1.Code != 0 || o1.Data.ID == 0 {
		t.Fatalf("first order failed status=%d code=%d id=%d", respO1.StatusCode, o1.Code, o1.Data.ID)
	}

	// 第二单
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab))
	req.Header.Set("Authorization", userAuth)
	req.Header.Set("Content-Type", "application/json")
	if _, err := http.DefaultClient.Do(req); err != nil {
		t.Fatalf("add cart 2 err: %v", err)
	}
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/orders/from-cart", bytes.NewReader(ob))
	req.Header.Set("Authorization", userAuth)
	req.Header.Set("Content-Type", "application/json")
	respO2, _ := http.DefaultClient.Do(req)
	var o2 struct {
		Code int
		Data struct{ ID uint }
	}
	json.NewDecoder(respO2.Body).Decode(&o2)
	respO2.Body.Close()
	if respO2.StatusCode != http.StatusOK || o2.Code != 0 || o2.Data.ID == 0 {
		t.Fatalf("second order failed status=%d code=%d id=%d", respO2.StatusCode, o2.Code, o2.Data.ID)
	}

	// 用户支付两单
	pay := func(id uint) {
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/orders/%d/pay", ts.URL, id), nil)
		req.Header.Set("Authorization", userAuth)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("pay err: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("pay status=%d", resp.StatusCode)
		}
		var pr struct{ Code int }
		if err := json.NewDecoder(resp.Body).Decode(&pr); err == nil && pr.Code != 0 {
			t.Fatalf("pay code=%d", pr.Code)
		}
	}
	pay(o1.Data.ID)
	pay(o2.Data.ID)

	shipAndComplete := func(id uint) {
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/orders/%d/deliver", ts.URL, id), nil)
		req.Header.Set("Authorization", adminAuth)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("deliver err: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			t.Fatalf("deliver status=%d", resp.StatusCode)
		}
		var dr struct{ Code int }
		json.NewDecoder(resp.Body).Decode(&dr)
		resp.Body.Close()
		if dr.Code != 0 {
			t.Fatalf("deliver code=%d", dr.Code)
		}
		req, _ = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/orders/%d/complete", ts.URL, id), nil)
		req.Header.Set("Authorization", adminAuth)
		resp2, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("complete err: %v", err)
		}
		if resp2.StatusCode != http.StatusOK {
			resp2.Body.Close()
			t.Fatalf("complete status=%d", resp2.StatusCode)
		}
		var cr struct{ Code int }
		json.NewDecoder(resp2.Body).Decode(&cr)
		resp2.Body.Close()
		if cr.Code != 0 {
			t.Fatalf("complete code=%d", cr.Code)
		}
	}
	shipAndComplete(o1.Data.ID)
	shipAndComplete(o2.Data.ID)

	// 查询统计
	url := fmt.Sprintf("%s/api/v1/admin/stores/%d/orders/stats", ts.URL, storeID)
	req, _ = http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", adminAuth)
	respStats, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("stats err: %v", err)
	}
	if respStats.StatusCode != 200 {
		t.Fatalf("stats status: %d", respStats.StatusCode)
	}
	var statsResp struct {
		Code int
		Data map[string]any
	}
	json.NewDecoder(respStats.Body).Decode(&statsResp)
	respStats.Body.Close()
	if statsResp.Code != 0 {
		t.Fatalf("stats code=%d", statsResp.Code)
	}
	if v, ok := statsResp.Data["total_orders"]; ok {
		switch n := v.(type) {
		case float64:
			if int(n) != 2 {
				t.Fatalf("total_orders != 2")
			}
		default:
			t.Fatalf("unexpected total_orders type")
		}
	} else {
		t.Fatalf("missing total_orders")
	}

	// completed_amount 可能以数字或字符串返回，这里兼容解析
	if v, ok := statsResp.Data["completed_amount"]; ok {
		switch s := v.(type) {
		case float64:
			if s < 19.99 || s > 20.01 {
				t.Fatalf("completed_amount not 20: %v", s)
			}
		case string:
			if s != "20" && s != "20.00" {
				t.Fatalf("completed_amount str not 20/20.00: %s", s)
			}
		default:
			t.Fatalf("unexpected completed_amount type")
		}
	} else {
		t.Fatalf("missing completed_amount")
	}
}
