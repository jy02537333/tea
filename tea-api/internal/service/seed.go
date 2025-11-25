package service

import (
	"fmt"

	"gorm.io/gorm"

	"tea-api/internal/model"
)

type SeedOptions struct {
	AssignOpenID string // optional: assign auditor role to this openid
}

// SeedRBAC seeds base permissions and an 'auditor' role with read permissions.
// It is safe to run multiple times (FirstOrCreate semantics).
func SeedRBAC(db *gorm.DB, opts SeedOptions) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	// permissions
	perms := []model.Permission{
		{BaseModel: model.BaseModel{UID: "perm-accrual-run"}, Name: "accrual:run", Module: "finance", Action: "run", Resource: "accrual"},
		{BaseModel: model.BaseModel{UID: "perm-accrual-export"}, Name: "accrual:export", Module: "finance", Action: "export", Resource: "accrual"},
		{BaseModel: model.BaseModel{UID: "perm-accrual-summary"}, Name: "accrual:summary", Module: "finance", Action: "summary", Resource: "accrual"},
		{BaseModel: model.BaseModel{UID: "perm-rbac-view"}, Name: "rbac:view", Module: "rbac", Action: "view", Resource: "*"},
		{BaseModel: model.BaseModel{UID: "perm-rbac-manage"}, Name: "rbac:manage", Module: "rbac", Action: "manage", Resource: "*"},
	}
	for i := range perms {
		_ = db.Where("name = ?", perms[i].Name).FirstOrCreate(&perms[i]).Error
	}

	// role auditor
	role := model.Role{BaseModel: model.BaseModel{UID: "role-auditor"}, Name: "auditor", DisplayName: "审计员"}
	_ = db.Where("name = ?", role.Name).FirstOrCreate(&role).Error

	// grant auditor basic view & summary
	// rbac:view
	var pView model.Permission
	_ = db.Where("name = ?", "rbac:view").First(&pView).Error
	var rpView model.RolePermission
	_ = db.Where("role_id = ? AND permission_id = ?", role.ID, pView.ID).
		FirstOrCreate(&rpView, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-auditor-rbac-view"}, RoleID: role.ID, PermissionID: pView.ID}).Error
	// accrual:summary
	var pSum model.Permission
	_ = db.Where("name = ?", "accrual:summary").First(&pSum).Error
	var rpSum model.RolePermission
	_ = db.Where("role_id = ? AND permission_id = ?", role.ID, pSum.ID).
		FirstOrCreate(&rpSum, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-auditor-accrual-summary"}, RoleID: role.ID, PermissionID: pSum.ID}).Error

	// optional: assign auditor to a user by openid
	if opts.AssignOpenID != "" {
		var u model.User
		if err := db.Where("open_id = ?", opts.AssignOpenID).First(&u).Error; err == nil && u.ID > 0 {
			var ur model.UserRole
			_ = db.Where("user_id = ? AND role_id = ?", u.ID, role.ID).
				FirstOrCreate(&ur, &model.UserRole{BaseModel: model.BaseModel{UID: "ur-assign-auditor"}, UserID: u.ID, RoleID: role.ID}).Error
		}
	}
	return nil
}
