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
		// Sprint C/SC6a: granular order permissions for admin routes
		{BaseModel: model.BaseModel{UID: "perm-order-accept"}, Name: "order:accept", Module: "order", Action: "accept", Resource: "*"},
		{BaseModel: model.BaseModel{UID: "perm-order-reject"}, Name: "order:reject", Module: "order", Action: "reject", Resource: "*"},
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

	// ensure admin role exists and grant order accept/reject
	adminRole := model.Role{BaseModel: model.BaseModel{UID: "role-admin"}, Name: "admin", DisplayName: "管理员"}
	_ = db.Where("name = ?", adminRole.Name).FirstOrCreate(&adminRole).Error
	// grant order:accept
	var pAccept model.Permission
	_ = db.Where("name = ?", "order:accept").First(&pAccept).Error
	var rpAccept model.RolePermission
	_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pAccept.ID).
		FirstOrCreate(&rpAccept, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-order-accept"}, RoleID: adminRole.ID, PermissionID: pAccept.ID}).Error
	// grant order:reject
	var pReject model.Permission
	_ = db.Where("name = ?", "order:reject").First(&pReject).Error
	var rpReject model.RolePermission
	_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pReject.ID).
		FirstOrCreate(&rpReject, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-order-reject"}, RoleID: adminRole.ID, PermissionID: pReject.ID}).Error

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
