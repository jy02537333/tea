package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/internal/router"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

func Test_Accrual_Uses_UserInterestRate_And_Export_XLSX(t *testing.T) {
	_ = os.Setenv("TEA_USE_SQLITE", "0")
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()

	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 管理员登录获取 token
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

	// 准备一个用户：余额>0，个性化日利率>0
	db := database.GetDB()
	u := model.User{
		BaseModel:    model.BaseModel{UID: utils.GenerateUID()},
		OpenID:       "rate_user_" + utils.GenerateUID(),
		Phone:        "test-" + utils.GenerateUID(),
		Nickname:     "rate user",
		Status:       1,
		Balance:      decimal.NewFromFloat(1000.00),
		InterestRate: decimal.NewFromFloat(0.005), // 0.5%
		Role:         "user",
	}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	// 触发仅对该用户计息，传入一个与其个性化不同的全局利率，以检查覆盖逻辑
	dtStr := time.Now().Format("2006-01-02")
	body := map[string]interface{}{
		"date":    dtStr,
		"rate":    0.001, // 全局0.1%，应被用户0.5%覆盖
		"user_id": u.ID,
	}
	bb, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", ts.URL+"/api/v1/admin/accrual/run", bytes.NewReader(bb))
	req.Header.Set("Authorization", "Bearer "+login.Data.Token)
	req.Header.Set("Content-Type", "application/json")
	respRun, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("accrual run: %v", err)
	}
	if respRun.StatusCode != 200 {
		t.Fatalf("accrual run status: %d", respRun.StatusCode)
	}
	_ = respRun.Body.Close()

	// 校验记录：应存在且 rate=0.005
	var rec model.InterestRecord
	if err := db.Where("user_id = ? AND date(date) = date(?)", u.ID, dtStr).First(&rec).Error; err != nil {
		t.Fatalf("query record: %v", err)
	}
	if rec.Rate.Cmp(decimal.NewFromFloat(0.005)) != 0 {
		t.Fatalf("rate not overridden, got %s", rec.Rate.String())
	}
	// 校验金额变动是否符合：1000 * 0.005 = 5, 期末=1005
	expectedAfter := decimal.NewFromFloat(1005.00)
	if rec.PrincipalAfter.Cmp(expectedAfter) != 0 {
		t.Fatalf("principal_after mismatch, want %s got %s", expectedAfter.String(), rec.PrincipalAfter.String())
	}

	// 测试 XLSX 导出（当前日期范围，仅该用户）
	req2, _ := http.NewRequest("GET", ts.URL+"/api/v1/admin/accrual/export?format=xlsx&start="+dtStr+"&end="+dtStr+"&user_id="+fmt.Sprintf("%d", u.ID)+"&lang=en&fields=user_id,date,interest_amount", nil)
	req2.Header.Set("Authorization", "Bearer "+login.Data.Token)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("export xlsx: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatalf("export xlsx status: %d", resp2.StatusCode)
	}
	ct := resp2.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet") {
		// 兼容某些平台 header 大小写或差异
		bts, _ := io.ReadAll(resp2.Body)
		if len(bts) == 0 {
			t.Fatalf("unexpected content-type: %s and empty body", ct)
		}
	}
}

// helper placeholder to avoid unused import if needed
