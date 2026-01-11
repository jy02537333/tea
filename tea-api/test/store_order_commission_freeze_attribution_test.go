package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

func Test_StoreOrderCommission_UsesFrozenSharer_NotOverriddenBinding(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	base := time.Now().UnixNano()
	sharerOpenID := fmt.Sprintf("sharer_openid_%d", base)
	buyerOpenID := fmt.Sprintf("buyer_openid_%d", base)
	overrideOpenID := fmt.Sprintf("override_openid_%d", base)

	_ = devLogin(t, ts, sharerOpenID)
	buyerToken := devLogin(t, ts, buyerOpenID)
	_ = devLogin(t, ts, overrideOpenID)

	buyerAuth := "Bearer " + buyerToken
	adminAuth := "Bearer " + devLogin(t, ts, "admin_openid")

	db := database.GetDB()
	var sharer, buyer, overrideUser model.User
	if err := db.Where("open_id = ?", sharerOpenID).First(&sharer).Error; err != nil {
		t.Fatalf("query sharer user: %v", err)
	}
	if err := db.Where("open_id = ?", buyerOpenID).First(&buyer).Error; err != nil {
		t.Fatalf("query buyer user: %v", err)
	}
	if err := db.Where("open_id = ?", overrideOpenID).First(&overrideUser).Error; err != nil {
		t.Fatalf("query override user: %v", err)
	}

	clearCart := func() {
		req, _ := http.NewRequest("DELETE", ts.URL+"/api/v1/cart/clear", nil)
		req.Header.Set("Authorization", buyerAuth)
		rc, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("clear cart err: %v", err)
		}
		defer rc.Body.Close()
		if rc.StatusCode != 200 {
			t.Fatalf("clear cart status: %d", rc.StatusCode)
		}
	}
	clearCart()

	createStore := func(name string) uint {
		st := map[string]any{"name": name, "address": "A", "latitude": 31.2, "longitude": 121.4, "status": 1}
		sb, _ := json.Marshal(st)
		req, _ := http.NewRequest("POST", ts.URL+"/api/v1/stores", bytes.NewReader(sb))
		req.Header.Set("Authorization", buyerAuth)
		req.Header.Set("Content-Type", "application/json")
		rs, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("create store err: %v", err)
		}
		defer rs.Body.Close()
		if rs.StatusCode != 200 {
			t.Fatalf("create store status: %d", rs.StatusCode)
		}
		var storeResp struct {
			Code int
			Data struct{ ID uint }
		}
		_ = json.NewDecoder(rs.Body).Decode(&storeResp)
		if storeResp.Code != 0 || storeResp.Data.ID == 0 {
			t.Fatalf("create store failed: %+v", storeResp)
		}
		return storeResp.Data.ID
	}
	storeID := createStore("门店-佣金冻结归属")

	// 创建分类
	catReq := map[string]any{"name": fmt.Sprintf("分类-%d", base)}
	cb, _ := json.Marshal(catReq)
	req, _ := http.NewRequest("POST", ts.URL+"/api/v1/categories", bytes.NewReader(cb))
	req.Header.Set("Authorization", buyerAuth)
	req.Header.Set("Content-Type", "application/json")
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create category err: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatalf("create category status: %d", resp2.StatusCode)
	}
	var catResp struct {
		Code int
		Data struct{ ID uint }
	}
	_ = json.NewDecoder(resp2.Body).Decode(&catResp)
	if catResp.Code != 0 || catResp.Data.ID == 0 {
		t.Fatalf("create category failed: %+v", catResp)
	}

	// 创建商品
	prodReq := map[string]any{"category_id": catResp.Data.ID, "name": fmt.Sprintf("商品-%d", base), "price": 10.00, "stock": 10, "status": 1}
	pb, _ := json.Marshal(prodReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/products", bytes.NewReader(pb))
	req.Header.Set("Authorization", buyerAuth)
	req.Header.Set("Content-Type", "application/json")
	resp3, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create product err: %v", err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != 200 {
		t.Fatalf("create product status: %d", resp3.StatusCode)
	}
	var prodResp struct {
		Code int
		Data struct{ ID uint }
	}
	_ = json.NewDecoder(resp3.Body).Decode(&prodResp)
	if prodResp.Code != 0 || prodResp.Data.ID == 0 {
		t.Fatalf("create product failed: %+v", prodResp)
	}

	// 绑定商品到门店库存
	bindPayload := map[string]any{"product_id": prodResp.Data.ID, "stock": 5, "price_override": "10"}
	bindBody, _ := json.Marshal(bindPayload)
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/admin/stores/%d/products", ts.URL, storeID), bytes.NewReader(bindBody))
	req.Header.Set("Authorization", adminAuth)
	req.Header.Set("Content-Type", "application/json")
	bindResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("assign product to store err: %v", err)
	}
	defer bindResp.Body.Close()
	if bindResp.StatusCode != 200 {
		t.Fatalf("assign product to store status: %d", bindResp.StatusCode)
	}

	// 加购
	add := map[string]any{"product_id": prodResp.Data.ID, "quantity": 1}
	ab, _ := json.Marshal(add)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab))
	req.Header.Set("Authorization", buyerAuth)
	req.Header.Set("Content-Type", "application/json")
	addResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("add cart err: %v", err)
	}
	addResp.Body.Close()

	// 门店下单，携带分享参数冻结 sharer
	orderReq := map[string]any{
		"delivery_type":   1,
		"store_id":        storeID,
		"sharer_uid":      sharer.ID,
		"share_store_id":  storeID,
		"order_type":      2,
		"address_info":    "{}",
		"remark":          "commission-freeze-attribution",
		"user_coupon_id":  0,
		"table_id":        0,
		"table_no":        "",
	}
	ob, _ := json.Marshal(orderReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/orders/from-cart", bytes.NewReader(ob))
	req.Header.Set("Authorization", buyerAuth)
	req.Header.Set("Content-Type", "application/json")
	orderResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create order err: %v", err)
	}
	defer orderResp.Body.Close()
	if orderResp.StatusCode != 200 {
		t.Fatalf("create order status: %d", orderResp.StatusCode)
	}
	var o struct {
		Code int
		Data struct{ ID uint }
	}
	_ = json.NewDecoder(orderResp.Body).Decode(&o)
	if o.Code != 0 || o.Data.ID == 0 {
		t.Fatalf("create order failed: %+v", o)
	}
	orderID := o.Data.ID

	// 支付
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/orders/%d/pay", ts.URL, orderID), nil)
	req.Header.Set("Authorization", buyerAuth)
	payResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("pay err: %v", err)
	}
	payResp.Body.Close()
	if payResp.StatusCode != 200 {
		t.Fatalf("pay status: %d", payResp.StatusCode)
	}

	// 覆盖绑定关系到另一个 referrer（PRD：最后一次点击覆盖）
	bindReq := map[string]any{"referrer_id": overrideUser.ID}
	bb, _ := json.Marshal(bindReq)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/referrals/bind", bytes.NewReader(bb))
	req.Header.Set("Authorization", buyerAuth)
	req.Header.Set("Content-Type", "application/json")
	bindOverrideResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("bind override err: %v", err)
	}
	bindOverrideResp.Body.Close()
	if bindOverrideResp.StatusCode != 200 {
		t.Fatalf("bind override status: %d", bindOverrideResp.StatusCode)
	}

	// 自取订单确认完成（触发佣金结算）
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/orders/%d/receive", ts.URL, orderID), nil)
	req.Header.Set("Authorization", buyerAuth)
	receiveResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("receive err: %v", err)
	}
	receiveResp.Body.Close()
	if receiveResp.StatusCode != 200 {
		t.Fatalf("receive status: %d", receiveResp.StatusCode)
	}

	// 直接查库：佣金应归属于下单时冻结的 sharer，而不是覆盖后的 referrer。
	var cms []model.Commission
	if err := db.Where("order_id = ? AND commission_type = ?", orderID, "direct").Find(&cms).Error; err != nil {
		t.Fatalf("query commissions err: %v", err)
	}
	if len(cms) != 1 {
		t.Fatalf("expected 1 direct commission, got %d", len(cms))
	}
	if cms[0].UserID != sharer.ID {
		t.Fatalf("commission user mismatch: want %d (sharer), got %d", sharer.ID, cms[0].UserID)
	}
	if cms[0].UserID == overrideUser.ID {
		t.Fatalf("commission should not follow overridden referrer")
	}
	if cms[0].Status != "available" {
		t.Fatalf("commission status mismatch: want available, got %s", cms[0].Status)
	}
	if cms[0].SourceUserID == nil || *cms[0].SourceUserID != buyer.ID {
		t.Fatalf("commission source_user_id mismatch: want %d (buyer), got %+v", buyer.ID, cms[0].SourceUserID)
	}
}
