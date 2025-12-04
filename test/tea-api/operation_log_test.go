//go:build ignore
// +build ignore

package test

import "testing"

func TestIgnoreOperationLog(t *testing.T) {
	t.Skip("Top-level consolidated tests are ignored when running from repo root. Run tests in submodules instead.")
}

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/internal/router"
	"tea-api/internal/service"
	"tea-api/pkg/database"
)

func Test_OperationLog_RBAC_And_Finance(t *testing.T) {
	// Use 测试环境2
	if os.Getenv("TEA_DSN") == "" {
		_ = os.Setenv("TEA_DSN", "root:gs963852@tcp(127.0.0.1:3306)/tea_shop?charset=utf8mb4&parseTime=True&loc=Local")
	}
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

	// 构造一个普通用户，方便 invalidate 接口
	u := model.User{BaseModel: model.BaseModel{UID: "u-oplog"}, OpenID: "oplog_openid", Phone: "oplog_phone", Nickname: "u", Status: 1}
	if err := db.Where("open_id = ?", u.OpenID).FirstOrCreate(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 登录 admin
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

	// 调用 RBAC 缓存失效接口
	invBody, _ := json.Marshal(map[string]any{"user_id": u.ID})
	reqInv, _ := http.NewRequest("POST", ts.URL+"/api/v1/admin/rbac/cache/invalidate", bytes.NewReader(invBody))
	reqInv.Header.Set("Authorization", "Bearer "+adminLogin.Data.Token)
	reqInv.Header.Set("Content-Type", "application/json")
	invResp, err := http.DefaultClient.Do(reqInv)
	if err != nil {
		t.Fatalf("invalidate: %v", err)
	}
	if invResp.StatusCode != 200 {
		t.Fatalf("invalidate status: %d", invResp.StatusCode)
	}
	_ = invResp.Body.Close()

	// 应写入一条 RBAC 模块的操作日志
	var cnt int64
	if err := db.Model(&model.OperationLog{}).Where("module = ? AND operation LIKE ?", "rbac", "%/rbac/cache/invalidate").Count(&cnt).Error; err != nil {
		t.Fatalf("query oplog: %v", err)
	}
	if cnt == 0 {
		t.Fatalf("expect rbac operation log written")
	}

	// 调用 finance accrual run
	reqRun, _ := http.NewRequest("POST", ts.URL+"/api/v1/admin/accrual/run", bytes.NewReader([]byte(`{"date":"2025-11-12","rate":0.001}`)))
	reqRun.Header.Set("Authorization", "Bearer "+adminLogin.Data.Token)
	reqRun.Header.Set("Content-Type", "application/json")
	runResp, err := http.DefaultClient.Do(reqRun)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if runResp.StatusCode != 200 {
		t.Fatalf("run status: %d", runResp.StatusCode)
	}
	_ = runResp.Body.Close()

	// 应写入一条 finance 模块的操作日志
	cnt = 0
	if err := db.Model(&model.OperationLog{}).Where("module = ? AND operation LIKE ?", "finance", "%/accrual/run").Count(&cnt).Error; err != nil {
		t.Fatalf("query oplog2: %v", err)
	}
	if cnt == 0 {
		t.Fatalf("expect finance operation log written")
	}
}
