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

func Test_Store_CRUD_And_List_With_Distance(t *testing.T) {
	_ = os.Setenv("TEA_USE_SQLITE", "0")
	_ = os.Setenv("TEA_SQLITE_PATH", "tea_test_store.db")
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 登录
	loginReq := map[string]string{"openid": "store_test_user"}
	b, _ := json.Marshal(loginReq)
	resp, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("dev-login err: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("login status: %d", resp.StatusCode)
	}
	var login struct {
		Code int
		Data struct{ Token string }
	}
	json.NewDecoder(resp.Body).Decode(&login)
	resp.Body.Close()
	if login.Code != 0 || login.Data.Token == "" {
		t.Fatalf("login failed")
	}
	auth := "Bearer " + login.Data.Token

	// 创建两个门店
	st1 := map[string]any{"name": "门店A", "address": "A路1号", "latitude": 31.2304, "longitude": 121.4737, "status": 1}
	st2 := map[string]any{"name": "门店B", "address": "B路2号", "latitude": 31.2200, "longitude": 121.4800, "status": 1}
	for _, st := range []map[string]any{st1, st2} {
		body, _ := json.Marshal(st)
		req, _ := http.NewRequest("POST", ts.URL+"/api/v1/stores", bytes.NewReader(body))
		req.Header.Set("Authorization", auth)
		req.Header.Set("Content-Type", "application/json")
		r1, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("create store err: %v", err)
		}
		if r1.StatusCode != 200 {
			t.Fatalf("create store status: %d", r1.StatusCode)
		}
		r1.Body.Close()
	}

	// 按指定坐标查询列表，返回含 distance_km 字段
	resp2, err := http.Get(ts.URL + "/api/v1/stores?lat=31.2250&lng=121.4750&page=1&limit=10")
	if err != nil {
		t.Fatalf("list stores err: %v", err)
	}
	if resp2.StatusCode != 200 {
		t.Fatalf("list stores status: %d", resp2.StatusCode)
	}
	var listResp struct {
		Code int
		Data []map[string]any
	}
	json.NewDecoder(resp2.Body).Decode(&listResp)
	resp2.Body.Close()
	if listResp.Code != 0 {
		t.Fatalf("list failed: %+v", listResp)
	}
	if len(listResp.Data) < 2 {
		t.Fatalf("store list size < 2")
	}
	_, ok := listResp.Data[0]["distance_km"]
	if !ok {
		t.Fatalf("distance_km missing in response")
	}
}
