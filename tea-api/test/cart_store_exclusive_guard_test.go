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

func Test_Cart_AddItem_StoreRole_ExclusiveAndNoMix(t *testing.T) {
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()
	db := database.GetDB()
	if db == nil {
		t.Fatalf("db is nil")
	}

	// Some environments might not have store_admins table migrated.
	// Create a minimal table for this test to exercise store-role guards.
	_ = db.Exec(`
CREATE TABLE IF NOT EXISTS store_admins (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  store_id BIGINT NOT NULL,
  created_at DATETIME NULL,
  updated_at DATETIME NULL,
  deleted_at DATETIME NULL
)`).Error

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 1) create a normal user via dev-login
	openid := "store_guard_user_openid"
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

	userAuth := login(openid)
	adminAuth := login("admin_openid")

	// Reset role to avoid contamination from prior runs.
	_ = db.Table("users").Where("open_id = ?", openid).Update("role", "user").Error
	_ = db.Exec("DELETE FROM store_admins WHERE user_id IN (SELECT id FROM users WHERE open_id = ?)", openid).Error
	userAuth = login(openid)

	// Clear cart as user to avoid historical data
	reqClr, _ := http.NewRequest("DELETE", ts.URL+"/api/v1/cart/clear", nil)
	reqClr.Header.Set("Authorization", userAuth)
	if respClr, err := http.DefaultClient.Do(reqClr); err == nil {
		respClr.Body.Close()
	}

	// 2) create a platform product P1 and add to cart (as user)
	catReq := map[string]any{"name": "guard_test_cat"}
	cb, _ := json.Marshal(catReq)
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

	prod1Req := map[string]any{"category_id": catResp.Data.ID, "name": "platform_p1", "price": 9.99, "stock": 10, "status": 1}
	pb1, _ := json.Marshal(prod1Req)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/products", bytes.NewReader(pb1))
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

	addP1 := map[string]any{"product_id": p1Resp.Data.ID, "quantity": 1}
	ab1, _ := json.Marshal(addP1)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab1))
	req.Header.Set("Authorization", userAuth)
	req.Header.Set("Content-Type", "application/json")
	respAdd1, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("add p1 err: %v", err)
	}
	respAdd1.Body.Close()
	if respAdd1.StatusCode != 200 {
		t.Fatalf("add p1 status: %d", respAdd1.StatusCode)
	}

	// 3) create a store S1
	stReq := map[string]any{"name": "guard_store", "address": "addr", "latitude": 31.2, "longitude": 121.4, "status": 1}
	sb, _ := json.Marshal(stReq)
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
	var sResp struct {
		Code int
		Data struct{ ID uint }
	}
	_ = json.NewDecoder(respS.Body).Decode(&sResp)
	respS.Body.Close()
	if sResp.Code != 0 || sResp.Data.ID == 0 {
		t.Fatalf("create store failed: %+v", sResp)
	}

	// 4) create exclusive product P2 and bind to store with biz_type=3
	prod2Req := map[string]any{"category_id": catResp.Data.ID, "name": "exclusive_p2", "price": 19.99, "stock": 10, "status": 1}
	pb2, _ := json.Marshal(prod2Req)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/products", bytes.NewReader(pb2))
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

	bindReq := map[string]any{"product_id": p2Resp.Data.ID, "stock": 5, "biz_type": 3, "price_override": "18.88"}
	bb, _ := json.Marshal(bindReq)
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/admin/stores/%d/products", ts.URL, sResp.Data.ID), bytes.NewReader(bb))
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

	// 5) flip role to store + bind store_admins, then relogin to get store-role token
	var userRow struct{ ID uint }
	if err := db.Table("users").Select("id").Where("open_id = ?", openid).Limit(1).Find(&userRow).Error; err != nil {
		t.Fatalf("query user id err: %v", err)
	}
	if userRow.ID == 0 {
		t.Fatalf("user id not found")
	}
	if err := db.Table("users").Where("id = ?", userRow.ID).Update("role", "store").Error; err != nil {
		t.Fatalf("update role err: %v", err)
	}
	// ensure store_admins mapping exists
	_ = db.Table("store_admins").Create(map[string]any{"user_id": userRow.ID, "store_id": sResp.Data.ID}).Error

	storeAuth := login(openid)

	// 6) adding exclusive P2 should be blocked because cart contains platform P1
	addP2 := map[string]any{"product_id": p2Resp.Data.ID, "quantity": 1}
	ab2, _ := json.Marshal(addP2)
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab2))
	req.Header.Set("Authorization", storeAuth)
	req.Header.Set("Content-Type", "application/json")
	respAdd2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("add p2 err: %v", err)
	}
	respAdd2.Body.Close()
	if respAdd2.StatusCode == 200 {
		t.Fatalf("expected mixed-cart guard, but got 200")
	}

	// 7) clear cart then add exclusive P2 should pass
	reqClr2, _ := http.NewRequest("DELETE", ts.URL+"/api/v1/cart/clear", nil)
	reqClr2.Header.Set("Authorization", storeAuth)
	respClr2, err := http.DefaultClient.Do(reqClr2)
	if err != nil {
		t.Fatalf("clear cart err: %v", err)
	}
	respClr2.Body.Close()
	if respClr2.StatusCode != 200 {
		t.Fatalf("clear cart status: %d", respClr2.StatusCode)
	}

	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab2))
	req.Header.Set("Authorization", storeAuth)
	req.Header.Set("Content-Type", "application/json")
	respAdd2b, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("add p2 after clear err: %v", err)
	}
	respAdd2b.Body.Close()
	if respAdd2b.StatusCode != 200 {
		t.Fatalf("add p2 after clear status: %d", respAdd2b.StatusCode)
	}

	// 8) adding platform P1 as store should be rejected
	req, _ = http.NewRequest("POST", ts.URL+"/api/v1/cart/items", bytes.NewReader(ab1))
	req.Header.Set("Authorization", storeAuth)
	req.Header.Set("Content-Type", "application/json")
	respAdd1b, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("add p1 as store err: %v", err)
	}
	respAdd1b.Body.Close()
	if respAdd1b.StatusCode == 200 {
		t.Fatalf("expected store-exclusive guard, but got 200")
	}
}
