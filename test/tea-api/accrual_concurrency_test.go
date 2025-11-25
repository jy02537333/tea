//go:build ignore
// +build ignore

package test

import "testing"

func TestIgnoreAccrualConcurrency(t *testing.T) {
	t.Skip("Top-level consolidated tests are ignored when running from repo root. Run tests in submodules instead.")
}

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/internal/router"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

// 并发下多请求触发同一天计息，应只插入一条记录并更新一次余额
func Test_Accrual_Concurrency_Idempotent(t *testing.T) {
	_ = os.Setenv("TEA_USE_SQLITE", "0")
	// Use 测试环境2
	_ = os.Setenv("TEA_DSN", "root:gs963852@tcp(127.0.0.1:3306)/tea_shop?charset=utf8mb4&parseTime=True&loc=Local")
	_ = os.Setenv("REDIS_ADDR", "127.0.0.1:6379")
	_ = os.Setenv("REDIS_PASS", "")
	_ = os.Setenv("RABBITMQ_ADDR", "amqp://guest:guest@127.0.0.1:5672/")
	// 使用独立的 SQLite 文件，避免旧表结构无唯一索引
	tmp := filepath.Join(os.TempDir(), "tea_accrual_concurrency_"+utils.GenerateUID()+".db")
	_ = os.Setenv("TEA_SQLITE_PATH", tmp)
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()

	db := database.GetDB()
	// 准备用户
	u := model.User{BaseModel: model.BaseModel{UID: utils.GenerateUID()}, OpenID: "u_" + utils.GenerateUID(), Phone: "p_" + utils.GenerateUID(), Nickname: "cc", Status: 1, Balance: decimal.NewFromFloat(1000)}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()
	// 管理员登录
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
		t.Fatalf("empty token")
	}

	date := time.Now().Format("2006-01-02")
	payload, _ := json.Marshal(map[string]any{"date": date, "rate": 0.01, "user_id": u.ID})

	// 并发 20 个请求
	wg := sync.WaitGroup{}
	errs := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("POST", ts.URL+"/api/v1/admin/accrual/run", bytes.NewReader(payload))
			req.Header.Set("Authorization", "Bearer "+login.Data.Token)
			req.Header.Set("Content-Type", "application/json")
			resp2, err := http.DefaultClient.Do(req)
			if err != nil {
				errs <- err
				return
			}
			_ = resp2.Body.Close()
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Fatalf("req error: %v", e)
	}

	// 断言只有一条记录
	var cnt int64
	if err := db.Model(&model.InterestRecord{}).Where("user_id = ? AND date(date) = date(?)", u.ID, date).Count(&cnt).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expect 1 record, got %d", cnt)
	}

	// 断言余额增加一次（1000 * 0.01 = 10）
	var uAfter model.User
	_ = db.First(&uAfter, u.ID).Error
	if uAfter.Balance.Cmp(decimal.NewFromFloat(1010)) != 0 {
		t.Fatalf("balance mismatch, got %s", uAfter.Balance.String())
	}
}
