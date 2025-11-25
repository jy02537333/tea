package test

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// 测试：缓存命中后，角色新增权限不会立刻生效；调用缓存失效接口后立即生效
func Test_RBAC_Cache_Invalidate(t *testing.T) {
	_ = os.Setenv("TEA_USE_SQLITE", "0")
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()
	db := database.GetDB()

	// seed 基础权限与角色
	if err := service.SeedRBAC(db, service.SeedOptions{}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// 创建 auditor 用户并绑定角色
	u := model.User{BaseModel: model.BaseModel{UID: "u-audit"}, OpenID: "audit_openid_1", Phone: "audit_phone_1", Nickname: "auditor", Status: 1}
	if err := db.Where("open_id = ?", u.OpenID).FirstOrCreate(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	var role model.Role
	_ = db.Where("name = ?", "auditor").First(&role).Error
	ur := model.UserRole{BaseModel: model.BaseModel{UID: "ur-audit"}, UserID: u.ID, RoleID: role.ID}
	_ = db.Where("user_id = ? AND role_id = ?", u.ID, role.ID).FirstOrCreate(&ur).Error

	// 确保起始状态下 auditor 没有 accrual:run 权限（避免上次测试残留导致首次即放行）
	var pRunInit model.Permission
	if err := db.Where("name = ?", "accrual:run").First(&pRunInit).Error; err == nil {
		_ = db.Unscoped().Where("role_id = ? AND permission_id = ?", role.ID, pRunInit.ID).Delete(&model.RolePermission{}).Error
	}

	// 启动服务
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

	// 审计员登录
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

	// 先访问需要 accrual:run 的接口，应被拒绝（auditor 只有 accrual:summary）
	reqRun1, _ := http.NewRequest("POST", ts.URL+"/api/v1/admin/accrual/run", bytes.NewReader([]byte(`{"date":"2025-11-12","rate":0.001}`)))
	reqRun1.Header.Set("Authorization", "Bearer "+uLogin.Data.Token)
	reqRun1.Header.Set("Content-Type", "application/json")
	runResp1, err := http.DefaultClient.Do(reqRun1)
	if err != nil {
		t.Fatalf("run1: %v", err)
	}
	if runResp1.StatusCode == 200 {
		t.Fatalf("expect forbidden before grant")
	}
	_ = runResp1.Body.Close()

	// 触发缓存构建：通过只读接口访问用户权限
	reqPerm, _ := http.NewRequest("GET", ts.URL+"/api/v1/admin/rbac/user-permissions?user_id="+fmt.Sprintf("%d", u.ID), nil)
	reqPerm.Header.Set("Authorization", "Bearer "+adminLogin.Data.Token)
	respPerm, err := http.DefaultClient.Do(reqPerm)
	if err != nil {
		t.Fatalf("user-permissions: %v", err)
	}
	_ = respPerm.Body.Close()

	// 给 auditor 角色新增 accrual:run 权限（DB 变更，但缓存仍旧）
	var pRun model.Permission
	_ = db.Where("name = ?", "accrual:run").First(&pRun).Error
	var rp model.RolePermission
	_ = db.Where("role_id = ? AND permission_id = ?", role.ID, pRun.ID).FirstOrCreate(&rp, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-auditor-run"}, RoleID: role.ID, PermissionID: pRun.ID}).Error

	// 直接再试一次，仍应被拒（缓存尚未失效）
	reqRun2, _ := http.NewRequest("POST", ts.URL+"/api/v1/admin/accrual/run", bytes.NewReader([]byte(`{"date":"2025-11-12","rate":0.001}`)))
	reqRun2.Header.Set("Authorization", "Bearer "+uLogin.Data.Token)
	reqRun2.Header.Set("Content-Type", "application/json")
	runResp2, err := http.DefaultClient.Do(reqRun2)
	if err != nil {
		t.Fatalf("run2: %v", err)
	}
	if runResp2.StatusCode == 200 {
		t.Fatalf("expect forbidden before invalidate")
	}
	_ = runResp2.Body.Close()

	// 管理员调用缓存失效接口
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

	// 再次尝试计息运行，应允许
	reqRun3, _ := http.NewRequest("POST", ts.URL+"/api/v1/admin/accrual/run", bytes.NewReader([]byte(`{"date":"2025-11-12","rate":0.001}`)))
	reqRun3.Header.Set("Authorization", "Bearer "+uLogin.Data.Token)
	reqRun3.Header.Set("Content-Type", "application/json")
	runResp3, err := http.DefaultClient.Do(reqRun3)
	if err != nil {
		t.Fatalf("run3: %v", err)
	}
	if runResp3.StatusCode != 200 {
		t.Fatalf("expect allowed after invalidate, got %d", runResp3.StatusCode)
	}
	_ = runResp3.Body.Close()
}
