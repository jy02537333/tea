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

func Test_Order_FromCart_NoMix_PlatformAndStore(t *testing.T) {
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	login := func(openid string) string {
		b, _ := json.Marshal(map[string]string{"openid": openid})
		resp, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatalf("dev-login err: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("dev-login status: %d", resp.StatusCode)
		}
		var out struct {
			Code int
			Data struct{ Token string }
		}
		_ = json.NewDecoder(resp.Body).Decode(&out)
		if out.Code != 0 || out.Data.Token == "" {
			t.Fatalf("login failed: %+v", out)
		}
		return "Bearer " + out.Data.Token
	}

	userAuth := login("mix_guard_user_openid")
	adminAuth := login("admin_openid")

	// clear cart
	reqClr, _ := http.NewRequest("DELETE", ts.URL+"/api/v1/cart/clear", nil)
	reqClr.Header.Set("Authorization", userAuth)
	if respClr, err := http.DefaultClient.Do(reqClr); err == nil {
		respClr.Body.Close()
	}

	// create category
	cb, _ := json.Marshal(map[string]any{"name": "mix_guard_cat"})
	req, _ := http.NewRequest("POST", ts.URL+"/api/v1/categories", bytes.NewReader(cb))
	req.Header.Set("Authorization", adminAuth)
	req.Header.Set("Content-Type", "application/json")
	respCat, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create category err: %v", err)
	}
	if respCat.StatusCode != 200 {
		t.Fatalf("create category status: %d", respCat.StatusCode)
	}
	var catResp struct {
		Code int
		Data struct{ ID uint }
	}
	_ = json.NewDecoder(respCat.Body).Decode(&catResp)
	respCat.Body.Close()
	if catResp.Code != 0 || catResp.Data.ID == 0 {
		t.Fatalf("create category failed: %+v", catResp)
	}

	// create store
	sb, _ := json.Marshal(map[string]any{"name": "mix_guard_store", "address": "addr", "latitude": 31.2, "longitude": 121.4, "status": 1})
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/stores", bytes.NewReader(sb))
	req.Header.Set("Authorization", adminAuth)
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
	_ = json.NewDecoder(respS.Body).Decode(&storeResp)
	respS.Body.Close()
	if storeResp.Code != 0 || storeResp.Data.ID == 0 {
		t.Fatalf("create store failed: %+v", storeResp)
	}

	// create platform product P1 (not bound to store)
	p1b, _ := json.Marshal(map[string]any{"category_id": catResp.Data.ID, "name": "mix_guard_platform", "price": 9.99, "stock": 10, "status": 1})
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/products", bytes.NewReader(p1b))
	req.Header.Set("Authorization", adminAuth)
	req.Header.Set("Content-Type", "application/json")
	respP1, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create product1 err: %v", err)
	}
	if respP1.StatusCode != 200 {
		t.Fatalf("create product1 status: %d", respP1.StatusCode)
	}
	var p1Resp struct {
		Code int
		Data struct{ ID uint }
	}
	_ = json.NewDecoder(respP1.Body).Decode(&p1Resp)
	respP1.Body.Close()
	if p1Resp.Code != 0 || p1Resp.Data.ID == 0 {
		t.Fatalf("create product1 failed: %+v", p1Resp)
	}

	// create store product P2 and bind to store
	p2b, _ := json.Marshal(map[string]any{"category_id": catResp.Data.ID, "name": "mix_guard_store_prod", "price": 19.99, "stock": 10, "status": 1})
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/products", bytes.NewReader(p2b))
	req.Header.Set("Authorization", adminAuth)
	req.Header.Set("Content-Type", "application/json")
	respP2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create product2 err: %v", err)
	}
	if respP2.StatusCode != 200 {
		t.Fatalf("create product2 status: %d", respP2.StatusCode)
	}
	var p2Resp struct {
		Code int
		Data struct{ ID uint }
	}
	_ = json.NewDecoder(respP2.Body).Decode(&p2Resp)
	respP2.Body.Close()
	if p2Resp.Code != 0 || p2Resp.Data.ID == 0 {
		t.Fatalf("create product2 failed: %+v", p2Resp)
	}

	bindReq := map[string]any{"product_id": p2Resp.Data.ID, "stock": 5, "price_override": "18.88"}
	bb, _ := json.Marshal(bindReq)
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/admin/stores/%d/products", ts.URL, storeResp.Data.ID), bytes.NewReader(bb))
	req.Header.Set("Authorization", adminAuth)
	req.Header.Set("Content-Type", "application/json")
	respBind, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("bind store product err: %v", err)
	}
	respBind.Body.Close()
	if respBind.StatusCode != 200 {
		t.Fatalf("bind store product status: %d", respBind.StatusCode)
	}

	// add both products to cart (user)
	add := func(pid uint) {
		ab, _ := json.Marshal(map[string]any{"product_id": pid, "quantity": 1})
		req, _ := http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab))
		req.Header.Set("Authorization", userAuth)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("add cart err: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("add cart status: %d", resp.StatusCode)
		}
	}
	add(p1Resp.Data.ID)
	add(p2Resp.Data.ID)

	// create store order from cart should fail due to mixed platform goods
	ob, _ := json.Marshal(map[string]any{"delivery_type": 1, "store_id": storeResp.Data.ID, "order_type": 2})
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/orders/from-cart", bytes.NewReader(ob))
	req.Header.Set("Authorization", userAuth)
	req.Header.Set("Content-Type", "application/json")
	respO, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create order err: %v", err)
	}
	respO.Body.Close()
	if respO.StatusCode == 200 {
		t.Fatalf("expected mixed-cart guard, but got 200")
	}
}
