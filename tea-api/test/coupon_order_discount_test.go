package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"tea-api/internal/config"
	"tea-api/internal/router"
	"tea-api/pkg/database"
)

func Test_Order_With_Coupon_Discount(t *testing.T) {
	// 确保每次运行使用全新数据库
	_ = os.Remove("tea_test_coupon.db")
	_ = os.Remove("tea_test_coupon.db-journal")
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 登录普通用户
	loginReq := map[string]string{"openid": "user_openid_coupon"}
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
		Data struct {
			Token string
			User  struct{ ID uint } `json:"user"`
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&login); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	resp.Body.Close()
	if login.Code != 0 || login.Data.Token == "" {
		t.Fatalf("login failed: %+v", login)
	}
	auth := "Bearer " + login.Data.Token
	userID := login.Data.User.ID

	// 管理员创建分类、商品
	catReq := map[string]any{"name": "券测试分类"}
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

	prodReq := map[string]any{"category_id": catResp.Data.ID, "name": "券测试商品", "price": 100.00, "stock": 5, "status": 1}
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

	// 创建满减券（满80减30）
	st := time.Now().Add(-time.Hour).Format(time.RFC3339)
	et := time.Now().Add(time.Hour).Format(time.RFC3339)
	couponReq := map[string]any{"name": "满减30", "type": 1, "amount": "30", "min_amount": "80", "total_count": 100, "status": 1, "start_time": st, "end_time": et}
	cb2, _ := json.Marshal(couponReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/coupons", bytes.NewReader(cb2))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	respC, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create coupon err: %v", err)
	}
	if respC.StatusCode != 200 {
		t.Fatalf("create coupon status: %d", respC.StatusCode)
	}
	var couponResp struct {
		Code int
		Data struct{ ID uint }
	}
	json.NewDecoder(respC.Body).Decode(&couponResp)
	respC.Body.Close()

	// 发券给当前用户
	grantReq := map[string]any{"coupon_id": couponResp.Data.ID, "user_id": userID}
	gb, _ := json.Marshal(grantReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/coupons/grant", bytes.NewReader(gb))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	respG, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("grant coupon err: %v", err)
	}
	if respG.StatusCode != 200 {
		t.Fatalf("grant coupon status: %d", respG.StatusCode)
	}
	var ucResp struct {
		Code int
		Data struct{ ID uint }
	}
	json.NewDecoder(respG.Body).Decode(&ucResp)
	respG.Body.Close()

	// 加入购物车并下单时使用优惠券
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

	orderReq := map[string]any{"delivery_type": 1, "user_coupon_id": ucResp.Data.ID}
	ob, _ := json.Marshal(orderReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/orders/from-cart", bytes.NewReader(ob))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	respO, _ := http.DefaultClient.Do(req)
	var orderResp struct {
		Code    int
		Message string
		Data    struct {
			ID             uint    `json:"id"`
			PayAmount      float64 `json:"pay_amount"`
			DiscountAmount float64 `json:"discount_amount"`
		}
	}
	json.NewDecoder(respO.Body).Decode(&orderResp)
	respO.Body.Close()
	if orderResp.Code != 0 || orderResp.Data.ID == 0 {
		t.Fatalf("create order failed: code=%d msg=%s", orderResp.Code, orderResp.Message)
	}

	// 价格校验：100 - 30 = 70
	if orderResp.Data.DiscountAmount < 29.99 || orderResp.Data.DiscountAmount > 30.01 {
		t.Fatalf("discount not 30, got %.2f", orderResp.Data.DiscountAmount)
	}
	if orderResp.Data.PayAmount < 69.99 || orderResp.Data.PayAmount > 70.01 {
		t.Fatalf("pay amount not 70, got %.2f", orderResp.Data.PayAmount)
	}

	// 用户可用券应减少；查询可用券
	req, _ = http.NewRequest("GET", ts.URL+"/api/v1/user/coupons", nil)
	req.Header.Set("Authorization", auth)
	respU, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("list user coupons err: %v", err)
	}
	if respU.StatusCode != 200 {
		t.Fatalf("list user coupons status: %d", respU.StatusCode)
	}
}
