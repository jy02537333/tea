//go:build ignore
// +build ignore

package test

import "testing"

func TestIgnoreOrderFlow(t *testing.T) {
	t.Skip("Top-level consolidated tests are ignored when running from repo root. Run tests in submodules instead.")
}

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

func Test_Order_Flow_Pay_Deliver_Complete(t *testing.T) {
	// TEA_USE_SQLITE removed; tests use MySQL by default
	// Use 测试环境2
	_ = os.Setenv("TEA_DSN", "root:gs963852@tcp(127.0.0.1:3306)/tea_shop?charset=utf8mb4&parseTime=True&loc=Local")
	_ = os.Setenv("REDIS_ADDR", "127.0.0.1:6379")
	_ = os.Setenv("REDIS_PASS", "")
	_ = os.Setenv("RABBITMQ_ADDR", "amqp://guest:guest@127.0.0.1:5672/")
	_ = os.Setenv("TEA_SQLITE_PATH", "tea_test_order_flow.db")
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 登录
	loginReq := map[string]string{"openid": "user_openid_flow"}
	b, _ := json.Marshal(loginReq)
	resp, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("dev-login request err: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("dev-login status: %d", resp.StatusCode)
	}
	var login struct {
		Code int
		Data struct{ Token string }
	}
	if err := json.NewDecoder(resp.Body).Decode(&login); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if login.Code != 0 || login.Data.Token == "" {
		t.Fatalf("login failed: %+v", login)
	}
	auth := "Bearer " + login.Data.Token

	// 准备商品
	catReq := map[string]any{"name": "流转分类"}
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

	prodReq := map[string]any{"category_id": catResp.Data.ID, "name": "流转商品", "price": 5.00, "stock": 5, "status": 1}
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

	// 加入购物车并下单
	addReq := map[string]any{"product_id": prodResp.Data.ID, "quantity": 1}
	ab, _ := json.Marshal(addReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	respAdd, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("add cart err: %v", err)
	}
	if respAdd.StatusCode != 200 {
		t.Fatalf("add cart status: %d", respAdd.StatusCode)
	}
	respAdd.Body.Close()

	orderReq := map[string]any{"delivery_type": 1}
	ob, _ := json.Marshal(orderReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/orders/from-cart", bytes.NewReader(ob))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	resp4, _ := http.DefaultClient.Do(req)
	var orderResp struct {
		Code    int
		Message string
		Data    struct{ ID uint }
	}
	json.NewDecoder(resp4.Body).Decode(&orderResp)
	resp4.Body.Close()
	if orderResp.Code != 0 || orderResp.Data.ID == 0 {
		t.Fatalf("create order failed: code=%d msg=%s", orderResp.Code, orderResp.Message)
	}

	// 支付
	req, _ = http.NewRequest("POST", fmt.Sprintf(ts.URL+"/api/v1/orders/%d/pay", orderResp.Data.ID), nil)
	req.Header.Set("Authorization", auth)
	resp5, _ := http.DefaultClient.Do(req)
	if resp5.StatusCode != 200 {
		var er struct {
			Code    int
			Message string
		}
		json.NewDecoder(resp5.Body).Decode(&er)
		resp5.Body.Close()
		t.Fatalf("pay status: %d, msg=%s", resp5.StatusCode, er.Message)
	}
	resp5.Body.Close()

	// 支付后尝试取消，应失败（400）
	req, _ = http.NewRequest("POST", fmt.Sprintf(ts.URL+"/api/v1/orders/%d/cancel", orderResp.Data.ID), nil)
	req.Header.Set("Authorization", auth)
	resp6, _ := http.DefaultClient.Do(req)
	if resp6.StatusCode == 200 {
		t.Fatalf("cancel should fail after pay")
	}
	resp6.Body.Close()

	// 发货与完成（需具备权限/管理员）。这里使用 admin_openid 登录执行。
	adminLoginReq := map[string]string{"openid": "admin_openid"}
	ab2, _ := json.Marshal(adminLoginReq)
	respA, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(ab2))
	if err != nil {
		t.Fatalf("admin dev-login err: %v", err)
	}
	var adminLogin struct {
		Code int
		Data struct{ Token string }
	}
	json.NewDecoder(respA.Body).Decode(&adminLogin)
	respA.Body.Close()
	if adminLogin.Code != 0 || adminLogin.Data.Token == "" {
		t.Fatalf("admin login failed")
	}
	adminAuth := "Bearer " + adminLogin.Data.Token

	req, _ = http.NewRequest("POST", fmt.Sprintf(ts.URL+"/api/v1/orders/%d/deliver", orderResp.Data.ID), nil)
	req.Header.Set("Authorization", adminAuth)
	resp7, _ := http.DefaultClient.Do(req)
	if resp7.StatusCode != 200 {
		t.Fatalf("deliver status: %d", resp7.StatusCode)
	}
	resp7.Body.Close()

	req, _ = http.NewRequest("POST", fmt.Sprintf(ts.URL+"/api/v1/orders/%d/complete", orderResp.Data.ID), nil)
	req.Header.Set("Authorization", adminAuth)
	resp8, _ := http.DefaultClient.Do(req)
	if resp8.StatusCode != 200 {
		t.Fatalf("complete status: %d", resp8.StatusCode)
	}
	resp8.Body.Close()
}
