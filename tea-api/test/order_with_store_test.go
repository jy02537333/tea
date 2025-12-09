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

func Test_Order_FromCart_With_Store(t *testing.T) {
	// sqlite env vars removed; tests use MySQL
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 登录
	loginReq := map[string]string{"openid": "user_openid_order_store"}
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

	// 清空购物车，避免历史数据影响
	req, _ := http.NewRequest("DELETE", ts.URL+"/api/v1/cart/clear", nil)
	req.Header.Set("Authorization", auth)
	if respClr, err := http.DefaultClient.Do(req); err == nil {
		respClr.Body.Close()
	} else {
		t.Fatalf("clear cart err: %v", err)
	}

	// 创建门店
	st := map[string]any{"name": "旗舰店", "address": "中心路8号", "latitude": 31.2304, "longitude": 121.4737, "status": 1}
	sb, _ := json.Marshal(st)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/stores", bytes.NewReader(sb))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	respS, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create store err: %v", err)
	}
	if respS.StatusCode != 200 {
		t.Fatalf("create store status: %d", respS.StatusCode)
	}
	var storeResp struct {
		Code int
		Data struct{ ID uint }
	}
	json.NewDecoder(respS.Body).Decode(&storeResp)
	respS.Body.Close()

	// 创建分类与商品
	catReq := map[string]any{"name": "门店下单分类"}
	cb, _ := json.Marshal(catReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/categories", bytes.NewReader(cb))
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

	prodReq := map[string]any{"category_id": catResp.Data.ID, "name": "门店专供", "price": 10.00, "stock": 3, "status": 1}
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

	// 绑定门店库存，确保门店可售该商品
	abAdmin, _ := json.Marshal(map[string]string{"openid": "admin_openid"})
	respAdmin, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(abAdmin))
	if err != nil {
		t.Fatalf("dev-login admin err: %v", err)
	}
	var adminLogin struct {
		Code int
		Data struct{ Token string }
	}
	json.NewDecoder(respAdmin.Body).Decode(&adminLogin)
	respAdmin.Body.Close()
	adminAuth := "Bearer " + adminLogin.Data.Token

	bindReq := map[string]any{"product_id": prodResp.Data.ID, "stock": 5, "price_override": "10.00"}
	bb, _ := json.Marshal(bindReq)
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/admin/stores/%d/products", ts.URL, storeResp.Data.ID), bytes.NewReader(bb))
	req.Header.Set("Authorization", adminAuth)
	req.Header.Set("Content-Type", "application/json")
	respBind, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("bind store inventory err: %v", err)
	}
	respBind.Body.Close()

	// 加入购物车
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

	// 下单并绑定门店，order_type=2(堂食)
	orderReq := map[string]any{"delivery_type": 1, "store_id": storeResp.Data.ID, "order_type": 2}
	ob, _ := json.Marshal(orderReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/orders/from-cart", bytes.NewReader(ob))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	respO, _ := http.DefaultClient.Do(req)
	var orderResp struct {
		Code    int
		Message string
		Data    struct {
			ID        uint
			StoreID   uint `json:"store_id"`
			OrderType int  `json:"order_type"`
		}
	}
	json.NewDecoder(respO.Body).Decode(&orderResp)
	respO.Body.Close()
	if orderResp.Code != 0 || orderResp.Data.ID == 0 {
		t.Fatalf("create order failed: code=%d msg=%s", orderResp.Code, orderResp.Message)
	}
	storeID := orderResp.Data.StoreID
	detailFetched := false
	var detailOrder struct {
		StoreID   uint
		OrderType int
	}
	fetchDetail := func() {
		if detailFetched {
			return
		}
		req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/v1/orders/%d", ts.URL, orderResp.Data.ID), nil)
		req.Header.Set("Authorization", auth)
		respDetail, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("fetch order detail err: %v", err)
		}
		var detailResp struct {
			Code int
			Data struct {
				Order struct {
					StoreID   uint `json:"store_id"`
					OrderType int  `json:"order_type"`
				} `json:"order"`
			}
		}
		json.NewDecoder(respDetail.Body).Decode(&detailResp)
		respDetail.Body.Close()
		if detailResp.Code != 0 {
			t.Fatalf("order detail failed: code=%d", detailResp.Code)
		}
		detailOrder.StoreID = detailResp.Data.Order.StoreID
		detailOrder.OrderType = detailResp.Data.Order.OrderType
		detailFetched = true
	}
	if storeID == 0 {
		// 兼容 API 暂未返回 store_id 的情况，查询订单详情确认
		fetchDetail()
		storeID = detailOrder.StoreID
	}
	if storeID != storeResp.Data.ID {
		t.Fatalf("store_id not match")
	}
	orderType := orderResp.Data.OrderType
	if orderType == 0 {
		fetchDetail()
		orderType = detailOrder.OrderType
	}
	if orderType != 2 {
		t.Fatalf("order_type not 2")
	}
}
