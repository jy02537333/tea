//go:build ignore
// +build ignore

package test

import "testing"

func TestIgnoreOrderAPI(t *testing.T) {
	t.Skip("Top-level consolidated tests are ignored when running from repo root. Run tests in submodules instead.")
}

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

func Test_Order_FromCart_List_Cancel(t *testing.T) {
	// sqlite env variables removed; tests use MySQL by default
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 登录
	loginReq := map[string]string{"openid": "user_openid_order"}
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

	// 创建分类
	catReq := map[string]any{"name": "下单分类"}
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
	if err := json.NewDecoder(resp2.Body).Decode(&catResp); err != nil {
		t.Fatalf("decode category: %v", err)
	}
	resp2.Body.Close()

	// 创建商品（库存10）
	prodReq := map[string]any{"category_id": catResp.Data.ID, "name": "下单商品", "price": "10.00", "stock": 10, "status": 1}
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
	if err := json.NewDecoder(resp3.Body).Decode(&prodResp); err != nil {
		t.Fatalf("decode product: %v", err)
	}
	resp3.Body.Close()

	// 加入购物车（2件）
	addReq := map[string]any{"product_id": prodResp.Data.ID, "quantity": 2}
	ab, _ := json.Marshal(addReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	resp4, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("add cart err: %v", err)
	}
	if resp4.StatusCode != 200 {
		t.Fatalf("add cart status: %d", resp4.StatusCode)
	}
	resp4.Body.Close()

	// 下单（自取）
	orderReq := map[string]any{"delivery_type": 1, "remark": "test"}
	ob, _ := json.Marshal(orderReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/orders/from-cart", bytes.NewReader(ob))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	resp5, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create order err: %v", err)
	}
	if resp5.StatusCode != 200 {
		t.Fatalf("create order status: %d", resp5.StatusCode)
	}
	var orderResp struct {
		Code int
		Data struct {
			ID      uint   `json:"id"`
			OrderNo string `json:"order_no"`
		}
	}
	if err := json.NewDecoder(resp5.Body).Decode(&orderResp); err != nil {
		t.Fatalf("decode order: %v", err)
	}
	resp5.Body.Close()
	if orderResp.Code != 0 || orderResp.Data.ID == 0 {
		t.Fatalf("invalid order resp: %+v", orderResp)
	}

	// 校验库存减少（列表查询商品）
	resp6, err := http.Get(ts.URL + "/api/v1/products?keyword=" + urlQuery("下单商品"))
	if err != nil {
		t.Fatalf("list products err: %v", err)
	}
	if resp6.StatusCode != 200 {
		t.Fatalf("list products status: %d", resp6.StatusCode)
	}
	var listResp struct {
		Code int
		Data []struct {
			ID    uint
			Stock int
		}
	}
	if err := json.NewDecoder(resp6.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode list products: %v", err)
	}
	resp6.Body.Close()
	if len(listResp.Data) == 0 {
		t.Fatalf("product not found in list")
	}
	if listResp.Data[0].Stock != 8 {
		t.Fatalf("stock not reduced to 8, got %d", listResp.Data[0].Stock)
	}

	// 取消订单
	req, _ = http.NewRequest("POST", fmt.Sprintf(ts.URL+"/api/v1/orders/%d/cancel", orderResp.Data.ID), nil)
	req.Header.Set("Authorization", auth)
	resp7, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("cancel order err: %v", err)
	}
	if resp7.StatusCode != 200 {
		t.Fatalf("cancel order status: %d", resp7.StatusCode)
	}
	resp7.Body.Close()

	// 再查库存应回到10
	resp8, err := http.Get(ts.URL + "/api/v1/products?keyword=" + urlQuery("下单商品"))
	if err != nil {
		t.Fatalf("list products err: %v", err)
	}
	if resp8.StatusCode != 200 {
		t.Fatalf("list products status: %d", resp8.StatusCode)
	}
	var listResp2 struct {
		Code int
		Data []struct {
			ID    uint
			Stock int
		}
	}
	if err := json.NewDecoder(resp8.Body).Decode(&listResp2); err != nil {
		t.Fatalf("decode list products: %v", err)
	}
	resp8.Body.Close()
	if len(listResp2.Data) == 0 {
		t.Fatalf("product not found in list2")
	}
	if listResp2.Data[0].Stock != 10 {
		t.Fatalf("stock not restored to 10, got %d", listResp2.Data[0].Stock)
	}
}

func urlQuery(s string) string { return s }
