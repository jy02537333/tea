package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tea-api/internal/config"
	"tea-api/internal/router"
	"tea-api/pkg/database"
)

func Test_AdminUsers_WithDevLogin(t *testing.T) {
	// TEA_USE_SQLITE deprecated; tests use MySQL by default

	// 加载配置（确保 System.Env=local 以允许 dev-login）
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}

	// 初始化数据库
	database.InitDatabase()

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// dev-login 管理员
	loginReq := map[string]string{"openid": "admin_openid"}
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
	if login.Data.Token == "" {
		t.Fatalf("empty token")
	}

	// 访问管理员接口
	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+login.Data.Token)
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("admin users request err: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatalf("admin users status: %d", resp2.StatusCode)
	}
}
