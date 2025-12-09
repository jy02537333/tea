package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"tea-api/internal/config"
	"tea-api/internal/router"
	"tea-api/pkg/database"
)

// 验证：绑定门店商品库存与价格覆盖后，下单会优先扣减门店库存并应用覆盖价
func Test_Store_Inventory_Order_Deduction(t *testing.T) {
	// TEA_USE_SQLITE/TEA_SQLITE_PATH deprecated; using MySQL
	// 清理旧的测试数据库，确保用例隔离
	_ = os.Remove("tea_test_store_inventory_order.db")
	_ = os.Remove("tea_test_store_inventory_order.db-journal")
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 普通用户登录
	b, _ := json.Marshal(map[string]string{"openid": "user_openid_store_inv"})
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
	userAuth := "Bearer " + login.Data.Token

	// 防止历史测试遗留的购物车/商品干扰当前用例
	req, _ := http.NewRequest("DELETE", ts.URL+"/api/v1/cart/clear", nil)
	req.Header.Set("Authorization", userAuth)
	if respClr, err := http.DefaultClient.Do(req); err == nil {
		respClr.Body.Close()
	} else {
		t.Fatalf("clear cart err: %v", err)
	}

	// 创建门店
	st := map[string]any{"name": "库存门店", "status": 1}
	sb, _ := json.Marshal(st)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/stores", bytes.NewReader(sb))
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

	// 创建分类与商品（商品价 12.00，库存10）
	cb, _ := json.Marshal(map[string]any{"name": "库存分类"})
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/categories", bytes.NewReader(cb))
	req.Header.Set("Authorization", userAuth)
	req.Header.Set("Content-Type", "application/json")
	respC, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create category err: %v", err)
	}
	var catResp struct {
		Code int
		Data struct{ ID uint }
	}
	json.NewDecoder(respC.Body).Decode(&catResp)
	respC.Body.Close()

	pb, _ := json.Marshal(map[string]any{"category_id": catResp.Data.ID, "name": "覆盖价商品", "price": 12.00, "stock": 10, "status": 1})
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/products", bytes.NewReader(pb))
	req.Header.Set("Authorization", userAuth)
	req.Header.Set("Content-Type", "application/json")
	respP, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create product err: %v", err)
	}
	var prodResp struct {
		Code int
		Data struct{ ID uint }
	}
	json.NewDecoder(respP.Body).Decode(&prodResp)
	respP.Body.Close()

	// 管理员绑定门店库存与覆盖价 8.50，库存=1
	abA, _ := json.Marshal(map[string]string{"openid": "admin_openid"})
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
	adminAuth := "Bearer " + adminLogin.Data.Token

	up := map[string]any{"product_id": prodResp.Data.ID, "stock": 1, "price_override": "8.50"}
	upb, _ := json.Marshal(up)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/admin/stores/"+jsonNumber(storeResp.Data.ID)+"/products", bytes.NewReader(upb))
	req.Header.Set("Authorization", adminAuth)
	req.Header.Set("Content-Type", "application/json")
	if _, err := http.DefaultClient.Do(req); err != nil {
		t.Fatalf("upsert inventory err: %v", err)
	}

	// 用户加入购物车并下单（绑定门店），应按8.50计价且门店库存扣为0
	ab, _ := json.Marshal(map[string]any{"product_id": prodResp.Data.ID, "quantity": 1})
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab))
	req.Header.Set("Authorization", userAuth)
	req.Header.Set("Content-Type", "application/json")
	if _, err := http.DefaultClient.Do(req); err != nil {
		t.Fatalf("add cart err: %v", err)
	}

	ob, _ := json.Marshal(map[string]any{"delivery_type": 1, "store_id": storeResp.Data.ID})
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/orders/from-cart", bytes.NewReader(ob))
	req.Header.Set("Authorization", userAuth)
	req.Header.Set("Content-Type", "application/json")
	respO, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create order err: %v", err)
	}
	var orderResp struct {
		Code int
		Data struct {
			ID        uint    `json:"id"`
			PayAmount float64 `json:"pay_amount"`
		}
	}
	json.NewDecoder(respO.Body).Decode(&orderResp)
	respO.Body.Close()
	if orderResp.Code != 0 {
		t.Fatalf("create order code=%d", orderResp.Code)
	}
	if orderResp.Data.PayAmount < 8.49 || orderResp.Data.PayAmount > 8.51 {
		t.Fatalf("pay_amount not override 8.50: %.2f", orderResp.Data.PayAmount)
	}

	// 再次下单，应因门店库存不足失败
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
	if respO2.StatusCode == 200 {
		var or2 struct {
			Code    int
			Message string
		}
		json.NewDecoder(respO2.Body).Decode(&or2)
		respO2.Body.Close()
		if or2.Code == 0 {
			t.Fatalf("second order unexpectedly succeeded")
		}
	}
}

// helper: convert uint to string via JSON to avoid fmt import noise above
func jsonNumber(u uint) string { b, _ := json.Marshal(u); return string(bytes.Trim(b, "\n")) }
