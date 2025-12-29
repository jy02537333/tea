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
		{BaseModel: model.BaseModel{UID: "perm-order-adjust"}, Name: "order:adjust", Module: "order", Action: "adjust", Resource: "order"},
		{BaseModel: model.BaseModel{UID: "perm-system-config-view"}, Name: "system:config:view", Module: "system", Action: "view", Resource: "config"},
		{BaseModel: model.BaseModel{UID: "perm-system-config-manage"}, Name: "system:config:manage", Module: "system", Action: "manage", Resource: "config"},
		{BaseModel: model.BaseModel{UID: "perm-marketing-banner-view"}, Name: "marketing:banner:view", Module: "marketing", Action: "view", Resource: "banner"},
		{BaseModel: model.BaseModel{UID: "perm-marketing-banner-manage"}, Name: "marketing:banner:manage", Module: "marketing", Action: "manage", Resource: "banner"},
		{BaseModel: model.BaseModel{UID: "perm-marketing-recharge-view"}, Name: "marketing:recharge:view", Module: "marketing", Action: "view", Resource: "recharge"},
		{BaseModel: model.BaseModel{UID: "perm-marketing-recharge-manage"}, Name: "marketing:recharge:manage", Module: "marketing", Action: "manage", Resource: "recharge"},
		{BaseModel: model.BaseModel{UID: "perm-user-partner-view"}, Name: "user:partner:view", Module: "user", Action: "view", Resource: "partner"},
		{BaseModel: model.BaseModel{UID: "perm-user-partner-manage"}, Name: "user:partner:manage", Module: "user", Action: "manage", Resource: "partner"},
	}
	for i := range perms {
		_ = db.Where("name = ?", perms[i].Name).FirstOrCreate(&perms[i]).Error
	}

	// Ensure default admin role exists and grant admin order:adjust by default.
	adminRole := model.Role{BaseModel: model.BaseModel{UID: "role-admin"}, Name: "admin", DisplayName: "管理员", Description: "系统管理员"}
	_ = db.Where("name = ?", adminRole.Name).FirstOrCreate(&adminRole).Error
	var pAdjust model.Permission
	_ = db.Where("name = ?", "order:adjust").First(&pAdjust).Error
	if adminRole.ID > 0 && pAdjust.ID > 0 {
		var rpAdminAdjust model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pAdjust.ID).
			FirstOrCreate(&rpAdminAdjust, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-order-adjust"}, RoleID: adminRole.ID, PermissionID: pAdjust.ID}).Error
	}
	var pSysView model.Permission
	_ = db.Where("name = ?", "system:config:view").First(&pSysView).Error
	if adminRole.ID > 0 && pSysView.ID > 0 {
		var rpAdminSysView model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pSysView.ID).
			FirstOrCreate(&rpAdminSysView, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-system-config-view"}, RoleID: adminRole.ID, PermissionID: pSysView.ID}).Error
	}
	var pSysManage model.Permission
	_ = db.Where("name = ?", "system:config:manage").First(&pSysManage).Error
	if adminRole.ID > 0 && pSysManage.ID > 0 {
		var rpAdminSysManage model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pSysManage.ID).
			FirstOrCreate(&rpAdminSysManage, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-system-config-manage"}, RoleID: adminRole.ID, PermissionID: pSysManage.ID}).Error
	}
	var pBannerView model.Permission
	_ = db.Where("name = ?", "marketing:banner:view").First(&pBannerView).Error
	if adminRole.ID > 0 && pBannerView.ID > 0 {
		var rpAdminBannerView model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pBannerView.ID).
			FirstOrCreate(&rpAdminBannerView, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-marketing-banner-view"}, RoleID: adminRole.ID, PermissionID: pBannerView.ID}).Error
	}
	var pBannerManage model.Permission
	_ = db.Where("name = ?", "marketing:banner:manage").First(&pBannerManage).Error
	if adminRole.ID > 0 && pBannerManage.ID > 0 {
		var rpAdminBannerManage model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pBannerManage.ID).
			FirstOrCreate(&rpAdminBannerManage, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-marketing-banner-manage"}, RoleID: adminRole.ID, PermissionID: pBannerManage.ID}).Error
	}
	var pRechargeView model.Permission
	_ = db.Where("name = ?", "marketing:recharge:view").First(&pRechargeView).Error
	if adminRole.ID > 0 && pRechargeView.ID > 0 {
		var rpAdminRechargeView model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pRechargeView.ID).
			FirstOrCreate(&rpAdminRechargeView, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-marketing-recharge-view"}, RoleID: adminRole.ID, PermissionID: pRechargeView.ID}).Error
	}
	var pRechargeManage model.Permission
	_ = db.Where("name = ?", "marketing:recharge:manage").First(&pRechargeManage).Error
	if adminRole.ID > 0 && pRechargeManage.ID > 0 {
		var rpAdminRechargeManage model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pRechargeManage.ID).
			FirstOrCreate(&rpAdminRechargeManage, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-marketing-recharge-manage"}, RoleID: adminRole.ID, PermissionID: pRechargeManage.ID}).Error
	}

	var pPartnerView model.Permission
	_ = db.Where("name = ?", "user:partner:view").First(&pPartnerView).Error
	if adminRole.ID > 0 && pPartnerView.ID > 0 {
		var rpAdminPartnerView model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pPartnerView.ID).
			FirstOrCreate(&rpAdminPartnerView, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-user-partner-view"}, RoleID: adminRole.ID, PermissionID: pPartnerView.ID}).Error
	}
	var pPartnerManage model.Permission
	_ = db.Where("name = ?", "user:partner:manage").First(&pPartnerManage).Error
	if adminRole.ID > 0 && pPartnerManage.ID > 0 {
		var rpAdminPartnerManage model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pPartnerManage.ID).
			FirstOrCreate(&rpAdminPartnerManage, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-user-partner-manage"}, RoleID: adminRole.ID, PermissionID: pPartnerManage.ID}).Error
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
