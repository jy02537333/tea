//go:build ignore
// +build ignore

package test

import "testing"

func TestIgnoreOperationLogListExport(t *testing.T) {
	t.Skip("Top-level consolidated tests are ignored when running from repo root. Run tests in submodules instead.")
}

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"tea-api/internal/config"
	"tea-api/internal/router"
	"tea-api/internal/service"
	"tea-api/pkg/database"
)

func Test_OperationLog_List_And_Export(t *testing.T) {
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
	db := database.GetDB()

	if err := service.SeedRBAC(db, service.SeedOptions{}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 管理员登录
	bAdmin, _ := json.Marshal(map[string]string{"openid": "admin_openid"})
	respAdm, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(bAdmin))
	if err != nil {
		t.Fatalf("dev-login admin: %v", err)
	}
	var adminLogin struct {
		Code int
		Data struct {
			Token string `json:"token"`
		}
	}
	_ = json.NewDecoder(respAdm.Body).Decode(&adminLogin)
	_ = respAdm.Body.Close()
	if adminLogin.Data.Token == "" {
		t.Fatalf("admin token empty")
	}

	// 触发一条操作日志：调用缓存失效接口（传一个不存在的用户ID也可）
	invBody, _ := json.Marshal(map[string]any{"user_id": 999999})
	reqInv, _ := http.NewRequest("POST", ts.URL+"/api/v1/admin/rbac/cache/invalidate", bytes.NewReader(invBody))
	reqInv.Header.Set("Authorization", "Bearer "+adminLogin.Data.Token)
	reqInv.Header.Set("Content-Type", "application/json")
	invResp, err := http.DefaultClient.Do(reqInv)
	if err != nil {
		t.Fatalf("invalidate: %v", err)
	}
	_ = invResp.Body.Close()

	// 列表查询
	reqList, _ := http.NewRequest("GET", ts.URL+"/api/v1/admin/logs/operations?module=rbac&page=1&limit=10", nil)
	reqList.Header.Set("Authorization", "Bearer "+adminLogin.Data.Token)
	listResp, err := http.DefaultClient.Do(reqList)
	if err != nil {
		t.Fatalf("logs list: %v", err)
	}
	if listResp.StatusCode != 200 {
		t.Fatalf("logs list status: %d", listResp.StatusCode)
	}
	_ = listResp.Body.Close()

	// 导出 CSV
	reqExp, _ := http.NewRequest("GET", ts.URL+"/api/v1/admin/logs/operations/export?module=rbac", nil)
	reqExp.Header.Set("Authorization", "Bearer "+adminLogin.Data.Token)
	expResp, err := http.DefaultClient.Do(reqExp)
	if err != nil {
		t.Fatalf("logs export: %v", err)
	}
	if expResp.StatusCode != 200 {
		t.Fatalf("logs export status: %d", expResp.StatusCode)
	}
	bs, _ := io.ReadAll(expResp.Body)
	_ = expResp.Body.Close()
	if len(bs) == 0 {
		t.Fatalf("export csv empty")
	}
}
