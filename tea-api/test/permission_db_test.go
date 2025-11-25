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
	"tea-api/pkg/database"
)

func Test_DBBacked_Permission_For_AccrualSummary(t *testing.T) {
	_ = os.Setenv("TEA_USE_SQLITE", "0")
	if err := config.LoadConfig("../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	database.InitDatabase()
	db := database.GetDB()

	// 准备角色、权限、用户及关联
	role := model.Role{BaseModel: model.BaseModel{UID: "role-auditor"}, Name: "auditor", DisplayName: "审计员"}
	_ = db.Where("name = ?", role.Name).FirstOrCreate(&role).Error
	perm := model.Permission{BaseModel: model.BaseModel{UID: "perm-accrual-summary"}, Name: "accrual:summary", Module: "finance", Action: "summary", Resource: "accrual"}
	_ = db.Where("name = ?", perm.Name).FirstOrCreate(&perm).Error
	// 角色-权限关联
	var rp model.RolePermission
	_ = db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).FirstOrCreate(&rp, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-auditor-accrual-summary"}, RoleID: role.ID, PermissionID: perm.ID}).Error

	// 创建用户并绑定角色
	u := model.User{BaseModel: model.BaseModel{UID: "user-auditor-uid"}, OpenID: "auditor_openid_123", Phone: "perm-" + "1234567890", Nickname: "auditor", Status: 1, Role: "user"}
	_ = db.Where("open_id = ?", u.OpenID).FirstOrCreate(&u).Error
	var ur model.UserRole
	_ = db.Where("user_id = ? AND role_id = ?", u.ID, role.ID).FirstOrCreate(&ur, &model.UserRole{BaseModel: model.BaseModel{UID: "ur-auditor"}, UserID: u.ID, RoleID: role.ID}).Error

	// 登录该用户
	r := router.SetupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()
	b, _ := json.Marshal(map[string]string{"openid": u.OpenID})
	resp, err := http.Post(ts.URL+"/api/v1/user/dev-login", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("dev-login auditor: %v", err)
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
		t.Fatalf("auditor token empty")
	}

	// 访问 accrual/summary 应该允许
	req1, _ := http.NewRequest("GET", ts.URL+"/api/v1/admin/accrual/summary", nil)
	req1.Header.Set("Authorization", "Bearer "+login.Data.Token)
	resp1, err := http.DefaultClient.Do(req1)
	if err != nil {
		t.Fatalf("accrual summary req: %v", err)
	}
	if resp1.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp1.StatusCode)
	}
	_ = resp1.Body.Close()

	// 访问仅管理员可用的 /admin/users 应该被拒绝
	req2, _ := http.NewRequest("GET", ts.URL+"/api/v1/admin/users", nil)
	req2.Header.Set("Authorization", "Bearer "+login.Data.Token)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("admin users req: %v", err)
	}
	if resp2.StatusCode == 200 {
		t.Fatalf("non-admin should not access /admin/users")
	}
	_ = resp2.Body.Close()
}
