//go:build ignore
// +build ignore

package test

import "testing"

func TestIgnoreCartAPI(t *testing.T) {
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

func Test_Cart_Add_List_Update_Remove(t *testing.T) {
	_ = os.Setenv("TEA_USE_SQLITE", "0")
	// Use 测试环境2
	if os.Getenv("TEA_DSN") == "" {
		_ = os.Setenv("TEA_DSN", "root:gs963852@tcp(127.0.0.1:3306)/tea_shop?charset=utf8mb4&parseTime=True&loc=Local")
	}
	_ = os.Setenv("REDIS_ADDR", "127.0.0.1:6379")
	_ = os.Setenv("REDIS_PASS", "")
	_ = os.Setenv("RABBITMQ_ADDR", "amqp://guest:guest@127.0.0.1:5672/")

	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// dev-login 普通用户
	loginReq := map[string]string{"openid": "user_openid_cart"}
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
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&login); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if login.Code != 0 {
		t.Fatalf("dev-login failed, code=%d", login.Code)
	}
	if login.Data.Token == "" {
		t.Fatalf("empty token")
	}
	authHeader := "Bearer " + login.Data.Token

	// 为了能添加购物车，需要先创建一个分类与商品（使用管理员）
	// 这里直接用同一用户创建商品，接口需要鉴权但不要求管理员角色
	// 1) 创建分类
	catReq := map[string]any{"name": "测试分类"}
	cb, _ := json.Marshal(catReq)
	req, _ := http.NewRequest("POST", ts.URL+"/api/v1/categories", bytes.NewReader(cb))
	req.Header.Set("Authorization", authHeader)
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
		Data struct {
			ID uint `json:"id"`
		}
	}
	if err := json.NewDecoder(resp2.Body).Decode(&catResp); err != nil {
		t.Fatalf("decode category: %v", err)
	}
	resp2.Body.Close()
	if catResp.Code != 0 || catResp.Data.ID == 0 {
		t.Fatalf("invalid category resp: %+v", catResp)
	}

	// 2) 创建商品
	prodReq := map[string]any{
		"category_id": catResp.Data.ID,
		"name":        "测试商品",
		"description": "desc",
		"images":      "[]",
		"price":       "9.90",
		"stock":       100,
		"status":      1,
	}
	pb, _ := json.Marshal(prodReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/products", bytes.NewReader(pb))
	req.Header.Set("Authorization", authHeader)
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
		Data struct {
			ID uint `json:"id"`
		}
	}
	if err := json.NewDecoder(resp3.Body).Decode(&prodResp); err != nil {
		t.Fatalf("decode product: %v", err)
	}
	resp3.Body.Close()
	if prodResp.Code != 0 || prodResp.Data.ID == 0 {
		t.Fatalf("invalid product resp: %+v", prodResp)
	}

	// 3) 添加到购物车
	addReq := map[string]any{"product_id": prodResp.Data.ID, "quantity": 2}
	ab, _ := json.Marshal(addReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	resp4, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("add cart err: %v", err)
	}
	if resp4.StatusCode != 200 {
		t.Fatalf("add cart status: %d", resp4.StatusCode)
	}
	var addResp struct {
		Code int
		Data struct {
			ID uint `json:"id"`
		}
	}
	if err := json.NewDecoder(resp4.Body).Decode(&addResp); err != nil {
		t.Fatalf("decode add cart: %v", err)
	}
	resp4.Body.Close()
	if addResp.Code != 0 || addResp.Data.ID == 0 {
		t.Fatalf("invalid add cart resp: %+v", addResp)
	}

	// 4) 查询购物车
	req, _ = http.NewRequest("GET", ts.URL+"/api/v1/cart", nil)
	req.Header.Set("Authorization", authHeader)
	resp5, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("list cart err: %v", err)
	}
	if resp5.StatusCode != 200 {
		t.Fatalf("list cart status: %d", resp5.StatusCode)
	}
	resp5.Body.Close()

	// 5) 更新数量
	upReq := map[string]any{"quantity": 3}
	ub, _ := json.Marshal(upReq)
	req, _ = http.NewRequest("PUT", ts.URL+"/api/v1/cart/items/"+fmt.Sprintf("%d", addResp.Data.ID), bytes.NewReader(ub))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	resp6, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("update cart err: %v", err)
	}
	if resp6.StatusCode != 200 {
		t.Fatalf("update cart status: %d", resp6.StatusCode)
	}
	resp6.Body.Close()

	// 6) 删除条目
	req, _ = http.NewRequest("DELETE", ts.URL+"/api/v1/cart/items/"+fmt.Sprintf("%d", addResp.Data.ID), nil)
	req.Header.Set("Authorization", authHeader)
	resp7, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("remove cart err: %v", err)
	}
	if resp7.StatusCode != 200 {
		t.Fatalf("remove cart status: %d", resp7.StatusCode)
	}
	resp7.Body.Close()
}
