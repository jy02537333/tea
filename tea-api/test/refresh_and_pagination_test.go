package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/internal/router"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

type loginResp struct {
	Code int `json:"code"`
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

type pageResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Total   int64           `json:"total"`
	Page    int             `json:"page"`
	Size    int             `json:"size"`
}

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	_ = os.Setenv("TEA_USE_SQLITE", "0")
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()
	r := router.SetupRouter()
	return httptest.NewServer(r)
}

func devLogin(t *testing.T, ts *httptest.Server, openid string) string {
	b, _ := json.Marshal(map[string]string{"openid": openid})
	resp, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("dev-login request err: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("dev-login status: %d", resp.StatusCode)
	}
	var lr loginResp
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if lr.Data.Token == "" {
		t.Fatalf("empty token")
	}
	return lr.Data.Token
}

func Test_RefreshToken_WorksAndKeepsRole(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// 以 admin_openid 登录
	token := devLogin(t, ts, "admin_openid")

	// 调用刷新接口
	req, _ := http.NewRequest("POST", ts.URL+"/api/v1/user/refresh", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("refresh request err: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("refresh status: %d", resp.StatusCode)
	}
	var lr loginResp
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		t.Fatalf("decode refresh: %v", err)
	}
	newToken := lr.Data.Token
	if newToken == "" {
		t.Fatalf("empty refreshed token")
	}
	if newToken == token {
		t.Log("refreshed token equals original; acceptable if same TTL policy, proceeding to verify access")
	}

	// 使用新token访问管理员接口应成功
	req2, _ := http.NewRequest("GET", ts.URL+"/api/v1/admin/users", nil)
	req2.Header.Set("Authorization", "Bearer "+newToken)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("admin users request err: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatalf("admin users status with refreshed token: %d", resp2.StatusCode)
	}
}

func Test_AdminUsers_PaginationEdges(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	adminToken := devLogin(t, ts, "admin_openid")
	// 直接向数据库创建若干普通用户，避免 phone 唯一索引冲突
	db := database.GetDB()
	base := time.Now().UnixNano() % 1_000_000_000
	for i := 1; i <= 3; i++ {
		u := &model.User{
			BaseModel: model.BaseModel{UID: utils.GenerateUID()},
			OpenID:    fmt.Sprintf("user_openid_%d_%03d", base, i),
			Phone:     fmt.Sprintf("18%09d%03d", base, i),
			Nickname:  fmt.Sprintf("U%03d", i),
			Status:    1,
			Role:      "user",
		}
		if err := db.Create(u).Error; err != nil {
			t.Fatalf("seed user %d: %v", i, err)
		}
	}

	// page=1, limit=1 应只返回1条，size回显1
	req1, _ := http.NewRequest("GET", ts.URL+"/api/v1/admin/users?page=1&limit=1", nil)
	req1.Header.Set("Authorization", "Bearer "+adminToken)
	resp1, err := http.DefaultClient.Do(req1)
	if err != nil {
		t.Fatalf("page1 req err: %v", err)
	}
	defer resp1.Body.Close()
	if resp1.StatusCode != 200 {
		t.Fatalf("page1 status: %d", resp1.StatusCode)
	}
	var p1 pageResp
	if err := json.NewDecoder(resp1.Body).Decode(&p1); err != nil {
		t.Fatalf("decode page1: %v", err)
	}
	if p1.Size != 1 {
		t.Fatalf("expect size=1, got %d", p1.Size)
	}
	var arr1 []map[string]interface{}
	_ = json.Unmarshal(p1.Data, &arr1)
	if len(arr1) != 1 {
		t.Fatalf("expect 1 item, got %d", len(arr1))
	}
	if p1.Total < 3 {
		t.Fatalf("expect total>=3, got %d", p1.Total)
	}

	// size=0：服务内部会按默认20查询，但响应会原样回显0
	req2, _ := http.NewRequest("GET", ts.URL+"/api/v1/admin/users?page=1&limit=0", nil)
	req2.Header.Set("Authorization", "Bearer "+adminToken)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("size0 req err: %v", err)
	}
	defer resp2.Body.Close()
	var p2 pageResp
	_ = json.NewDecoder(resp2.Body).Decode(&p2)
	if p2.Size != 0 {
		t.Fatalf("expect size echo=0 when limit=0, got %d", p2.Size)
	}
	var arr2 []map[string]interface{}
	_ = json.Unmarshal(p2.Data, &arr2)
	if len(arr2) == 0 {
		t.Fatalf("expect non-empty data when there are users")
	}

	// page 很大时应返回空列表
	req3, _ := http.NewRequest("GET", ts.URL+"/api/v1/admin/users?page=1000&limit=20", nil)
	req3.Header.Set("Authorization", "Bearer "+adminToken)
	resp3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatalf("page large req err: %v", err)
	}
	defer resp3.Body.Close()
	var p3 pageResp
	_ = json.NewDecoder(resp3.Body).Decode(&p3)
	var arr3 []map[string]interface{}
	_ = json.Unmarshal(p3.Data, &arr3)
	if len(arr3) != 0 {
		t.Fatalf("expect empty list on large page, got %d", len(arr3))
	}
}
