package test

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

// 验证：批量给角色新增多个权限后，服务自动对该角色下用户的权限缓存失效，多个权限立即生效
func Test_RBAC_BulkAssign_AutoInvalidate(t *testing.T) {
	// TEA_USE_SQLITE removed; tests use MySQL by default
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

	// 找到 auditor 角色 与 accrual:run/export 权限
	var role model.Role
	if err := db.Where("name = ?", "auditor").First(&role).Error; err != nil {
		t.Fatalf("role: %v", err)
	}
	var pRun, pExport model.Permission
	if err := db.Where("name = ?", "accrual:run").First(&pRun).Error; err != nil {
		t.Fatalf("perm run: %v", err)
	}
	if err := db.Where("name = ?", "accrual:export").First(&pExport).Error; err != nil {
		t.Fatalf("perm export: %v", err)
	}

	// 清理角色上的 run/export 权限，准备起点
	_ = db.Unscoped().Where("role_id = ? AND permission_id IN ?", role.ID, []uint{pRun.ID, pExport.ID}).Delete(&model.RolePermission{}).Error

	// 创建一个审计员用户并绑定角色
	u := model.User{BaseModel: model.BaseModel{UID: "u-bulk-inv"}, OpenID: "bulk_inv_openid", Phone: "bulk_inv_phone", Nickname: "aud2", Status: 1}
	if err := db.Where("open_id = ?", u.OpenID).FirstOrCreate(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	var ur model.UserRole
	_ = db.Where("user_id = ? AND role_id = ?", u.ID, role.ID).FirstOrCreate(&ur, &model.UserRole{BaseModel: model.BaseModel{UID: "ur-bulk-inv"}, UserID: u.ID, RoleID: role.ID}).Error

	// 启动服务
	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	// 登录 admin 和审计员
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

	bU, _ := json.Marshal(map[string]string{"openid": u.OpenID})
	respU, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(bU))
	if err != nil {
		t.Fatalf("dev-login user: %v", err)
	}
	var uLogin struct {
		Code int
		Data struct {
			Token string `json:"token"`
		}
	}
	_ = json.NewDecoder(respU.Body).Decode(&uLogin)
	_ = respU.Body.Close()
	if uLogin.Data.Token == "" {
		t.Fatalf("user token empty")
	}

	// 尝试 run / export，均应被拒
	reqRun1, _ := http.NewRequest("POST", ts.URL+"/api/v1/admin/accrual/run", bytes.NewReader([]byte(`{"date":"2025-11-12","rate":0.001}`)))
	reqRun1.Header.Set("Authorization", "Bearer "+uLogin.Data.Token)
	reqRun1.Header.Set("Content-Type", "application/json")
	runResp1, err := http.DefaultClient.Do(reqRun1)
	if err != nil {
		t.Fatalf("run1: %v", err)
	}
	if runResp1.StatusCode == 200 {
		t.Fatalf("expect forbidden before assign")
	}
	_ = runResp1.Body.Close()

	reqExp1, _ := http.NewRequest("GET", ts.URL+"/api/v1/admin/accrual/export", nil)
	reqExp1.Header.Set("Authorization", "Bearer "+uLogin.Data.Token)
	expResp1, err := http.DefaultClient.Do(reqExp1)
	if err != nil {
		t.Fatalf("export1: %v", err)
	}
	if expResp1.StatusCode == 200 {
		t.Fatalf("expect forbidden before assign (export)")
	}
	_ = expResp1.Body.Close()

	// 构建只读缓存
	reqPerm, _ := http.NewRequest("GET", ts.URL+"/api/v1/admin/rbac/user-permissions?user_id="+fmtID(u.ID), nil)
	reqPerm.Header.Set("Authorization", "Bearer "+adminLogin.Data.Token)
	if _, err := http.DefaultClient.Do(reqPerm); err != nil {
		t.Fatalf("user-permissions: %v", err)
	}

	// 批量授权：run + export
	bodyAssign, _ := json.Marshal(map[string]any{"role_id": role.ID, "permission_ids": []uint{pRun.ID, pExport.ID}})
	reqAssign, _ := http.NewRequest("POST", ts.URL+"/api/v1/admin/rbac/role/assign-permissions", bytes.NewReader(bodyAssign))
	reqAssign.Header.Set("Authorization", "Bearer "+adminLogin.Data.Token)
	reqAssign.Header.Set("Content-Type", "application/json")
	respAssign, err := http.DefaultClient.Do(reqAssign)
	if err != nil {
		t.Fatalf("assign bulk: %v", err)
	}
	if respAssign.StatusCode != 200 {
		t.Fatalf("assign status: %d", respAssign.StatusCode)
	}
	_ = respAssign.Body.Close()

	// 再次访问：run 应放行
	reqRun2, _ := http.NewRequest("POST", ts.URL+"/api/v1/admin/accrual/run", bytes.NewReader([]byte(`{"date":"2025-11-12","rate":0.001}`)))
	reqRun2.Header.Set("Authorization", "Bearer "+uLogin.Data.Token)
	reqRun2.Header.Set("Content-Type", "application/json")
	runResp2, err := http.DefaultClient.Do(reqRun2)
	if err != nil {
		t.Fatalf("run2: %v", err)
	}
	if runResp2.StatusCode != 200 {
		t.Fatalf("expect allowed after bulk assign (run), got %d", runResp2.StatusCode)
	}
	_ = runResp2.Body.Close()

	// 再次访问：export 应放行
	reqExp2, _ := http.NewRequest("GET", ts.URL+"/api/v1/admin/accrual/export", nil)
	reqExp2.Header.Set("Authorization", "Bearer "+uLogin.Data.Token)
	expResp2, err := http.DefaultClient.Do(reqExp2)
	if err != nil {
		t.Fatalf("export2: %v", err)
	}
	if expResp2.StatusCode != 200 {
		t.Fatalf("expect allowed after bulk assign (export), got %d", expResp2.StatusCode)
	}
	_ = expResp2.Body.Close()
}
