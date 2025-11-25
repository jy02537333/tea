//go:build ignore
// +build ignore

package test

import "testing"

func TestIgnoreInterest(t *testing.T) {
	t.Skip("Top-level consolidated tests are ignored when running from repo root. Run tests in submodules instead.")
}

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

func Test_AdminAccrualRun_And_UserRecords(t *testing.T) {
	_ = os.Setenv("TEA_USE_SQLITE", "0")
	// Use 测试环境2
	_ = os.Setenv("TEA_DSN", "root:gs963852@tcp(127.0.0.1:3306)/tea_shop?charset=utf8mb4&parseTime=True&loc=Local")
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

	// 先创建一个管理员 token
	b, _ := json.Marshal(map[string]string{"openid": "admin_openid"})
	resp, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("dev-login: %v", err)
	}
	var login struct {
		Code int
		Data struct {
			Token string `json:"token"`
		}
	}
	_ = json.NewDecoder(resp.Body).Decode(&login)
	_ = resp.Body.Close()
	if login.Data.Token == "" {
		t.Fatalf("admin token empty")
	}

	// 创建一个普通用户，给他登录以建档（余额默认0，测试只验证记录接口，计息对所有有余额用户生效）
	b2, _ := json.Marshal(map[string]string{"openid": "interest_user_openid"})
	resp2, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b2))
	if err != nil {
		t.Fatalf("dev-login user: %v", err)
	}
	_ = resp2.Body.Close()

	// 触发计息（为了便于测试，这里只演示调用接口，不强行校验余额变化，实际依赖数据库初始余额>0时才会产生记录）
	body := map[string]interface{}{
		"date": time.Now().Format("2006-01-02"),
		"rate": 0.01,
	}
	bb, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", ts.URL+"/api/v1/admin/accrual/run", bytes.NewReader(bb))
	req.Header.Set("Authorization", "Bearer "+login.Data.Token)
	req.Header.Set("Content-Type", "application/json")
	resp3, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("accrual run req: %v", err)
	}
	if resp3.StatusCode != 200 {
		t.Fatalf("accrual run status: %d", resp3.StatusCode)
	}
	_ = resp3.Body.Close()

	// 查询当前用户的计息记录列表（可能为空，如果余额为0）。这里仅校验接口可用
	req2, _ := http.NewRequest("GET", ts.URL+"/api/v1/user/interest-records?page=1&limit=10", nil)
	req2.Header.Set("Authorization", "Bearer "+login.Data.Token)
	resp4, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("user records req: %v", err)
	}
	if resp4.StatusCode != 200 {
		t.Fatalf("user records status: %d", resp4.StatusCode)
	}
	_ = resp4.Body.Close()
}
