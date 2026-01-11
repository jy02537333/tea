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
		{BaseModel: model.BaseModel{UID: "perm-store-excl-view"}, Name: "store:exclusive_products:view", Module: "store", Action: "view", Resource: "exclusive_products"},
		{BaseModel: model.BaseModel{UID: "perm-store-excl-manage"}, Name: "store:exclusive_products:manage", Module: "store", Action: "manage", Resource: "exclusive_products"},
		{BaseModel: model.BaseModel{UID: "perm-system-config-view"}, Name: "system:config:view", Module: "system", Action: "view", Resource: "config"},
		{BaseModel: model.BaseModel{UID: "perm-system-config-manage"}, Name: "system:config:manage", Module: "system", Action: "manage", Resource: "config"},
		{BaseModel: model.BaseModel{UID: "perm-marketing-banner-view"}, Name: "marketing:banner:view", Module: "marketing", Action: "view", Resource: "banner"},
		{BaseModel: model.BaseModel{UID: "perm-marketing-banner-manage"}, Name: "marketing:banner:manage", Module: "marketing", Action: "manage", Resource: "banner"},
		{BaseModel: model.BaseModel{UID: "perm-marketing-recharge-view"}, Name: "marketing:recharge:view", Module: "marketing", Action: "view", Resource: "recharge"},
		{BaseModel: model.BaseModel{UID: "perm-marketing-recharge-manage"}, Name: "marketing:recharge:manage", Module: "marketing", Action: "manage", Resource: "recharge"},
		{BaseModel: model.BaseModel{UID: "perm-user-partner-view"}, Name: "user:partner:view", Module: "user", Action: "view", Resource: "partner"},
		{BaseModel: model.BaseModel{UID: "perm-user-partner-manage"}, Name: "user:partner:manage", Module: "user", Action: "manage", Resource: "partner"},
		// Store wallet & finance view (for store admins)
		{BaseModel: model.BaseModel{UID: "perm-store-wallet-view"}, Name: "store:wallet:view", Module: "store", Action: "view", Resource: "wallet"},
		// Store accounts & withdraw permissions
		{BaseModel: model.BaseModel{UID: "perm-store-accounts-view"}, Name: "store:accounts:view", Module: "store", Action: "view", Resource: "accounts"},
		{BaseModel: model.BaseModel{UID: "perm-store-accounts-manage"}, Name: "store:accounts:manage", Module: "store", Action: "manage", Resource: "accounts"},
		{BaseModel: model.BaseModel{UID: "perm-store-withdraw-view"}, Name: "store:withdraw:view", Module: "store", Action: "view", Resource: "withdraw"},
		{BaseModel: model.BaseModel{UID: "perm-store-withdraw-apply"}, Name: "store:withdraw:apply", Module: "store", Action: "apply", Resource: "withdraw"},
		// Store coupons & activities permissions
		{BaseModel: model.BaseModel{UID: "perm-store-coupons-view"}, Name: "store:coupons:view", Module: "store", Action: "view", Resource: "coupons"},
		{BaseModel: model.BaseModel{UID: "perm-store-coupons-manage"}, Name: "store:coupons:manage", Module: "store", Action: "manage", Resource: "coupons"},
		{BaseModel: model.BaseModel{UID: "perm-store-activities-view"}, Name: "store:activities:view", Module: "store", Action: "view", Resource: "activities"},
		{BaseModel: model.BaseModel{UID: "perm-store-activities-manage"}, Name: "store:activities:manage", Module: "store", Action: "manage", Resource: "activities"},
	}
	for i := range perms {
		_ = db.Where("name = ?", perms[i].Name).FirstOrCreate(&perms[i]).Error
	}

	// Ensure default admin role exists and grant admin order:adjust by default.
	adminRole := model.Role{BaseModel: model.BaseModel{UID: "role-admin"}, Name: "admin", DisplayName: "管理员", Description: "系统管理员"}
	_ = db.Where("name = ?", adminRole.Name).FirstOrCreate(&adminRole).Error

	// Ensure store role exists (for store admins)
	storeRole := model.Role{BaseModel: model.BaseModel{UID: "role-store"}, Name: "store", DisplayName: "门店管理员", Description: "门店后台管理员"}
	_ = db.Where("name = ?", storeRole.Name).FirstOrCreate(&storeRole).Error
	var pAdjust model.Permission
	_ = db.Where("name = ?", "order:adjust").First(&pAdjust).Error
	if adminRole.ID > 0 && pAdjust.ID > 0 {
		var rpAdminAdjust model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pAdjust.ID).
			FirstOrCreate(&rpAdminAdjust, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-order-adjust"}, RoleID: adminRole.ID, PermissionID: pAdjust.ID}).Error
	}
	var pStoreExclusiveView model.Permission
	_ = db.Where("name = ?", "store:exclusive_products:view").First(&pStoreExclusiveView).Error
	if adminRole.ID > 0 && pStoreExclusiveView.ID > 0 {
		var rpAdminStoreExclusiveView model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pStoreExclusiveView.ID).
			FirstOrCreate(&rpAdminStoreExclusiveView, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-store-excl-view"}, RoleID: adminRole.ID, PermissionID: pStoreExclusiveView.ID}).Error
	}
	var pStoreExclusiveManage model.Permission
	_ = db.Where("name = ?", "store:exclusive_products:manage").First(&pStoreExclusiveManage).Error
	if adminRole.ID > 0 && pStoreExclusiveManage.ID > 0 {
		var rpAdminStoreExclusiveManage model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pStoreExclusiveManage.ID).
			FirstOrCreate(&rpAdminStoreExclusiveManage, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-store-excl-manage"}, RoleID: adminRole.ID, PermissionID: pStoreExclusiveManage.ID}).Error
	}

	// Grant store role permissions for exclusive products (view/manage)
	if storeRole.ID > 0 && pStoreExclusiveView.ID > 0 {
		var rpStoreView model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", storeRole.ID, pStoreExclusiveView.ID).
			FirstOrCreate(&rpStoreView, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-store-excl-view"}, RoleID: storeRole.ID, PermissionID: pStoreExclusiveView.ID}).Error
	}
	if storeRole.ID > 0 && pStoreExclusiveManage.ID > 0 {
		var rpStoreManage model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", storeRole.ID, pStoreExclusiveManage.ID).
			FirstOrCreate(&rpStoreManage, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-store-excl-manage"}, RoleID: storeRole.ID, PermissionID: pStoreExclusiveManage.ID}).Error
	}

	// Ensure all users in store_admins are assigned 'store' role (idempotent)
	if storeRole.ID > 0 {
		type storeAdminRow struct{ UserID uint }
		var rows []storeAdminRow
		_ = db.Table("store_admins").Select("user_id").Find(&rows).Error
		for _, r := range rows {
			if r.UserID == 0 {
				continue
			}
			var ur model.UserRole
			_ = db.Where("user_id = ? AND role_id = ?", r.UserID, storeRole.ID).
				FirstOrCreate(&ur, &model.UserRole{BaseModel: model.BaseModel{UID: fmt.Sprintf("ur-store-%d", r.UserID)}, UserID: r.UserID, RoleID: storeRole.ID}).Error
		}
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

	// Grant store wallet view to store role (and admin for completeness)
	var pStoreWalletView model.Permission
	_ = db.Where("name = ?", "store:wallet:view").First(&pStoreWalletView).Error
	if storeRole.ID > 0 && pStoreWalletView.ID > 0 {
		var rpStoreWalletView model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", storeRole.ID, pStoreWalletView.ID).
			FirstOrCreate(&rpStoreWalletView, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-store-wallet-view"}, RoleID: storeRole.ID, PermissionID: pStoreWalletView.ID}).Error
	}
	if adminRole.ID > 0 && pStoreWalletView.ID > 0 {
		var rpAdminStoreWalletView model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", adminRole.ID, pStoreWalletView.ID).
			FirstOrCreate(&rpAdminStoreWalletView, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-admin-store-wallet-view"}, RoleID: adminRole.ID, PermissionID: pStoreWalletView.ID}).Error
	}

	// Grant store accounts view/manage to store role
	var pStoreAccountsView model.Permission
	_ = db.Where("name = ?", "store:accounts:view").First(&pStoreAccountsView).Error
	if storeRole.ID > 0 && pStoreAccountsView.ID > 0 {
		var rpStoreAccountsView model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", storeRole.ID, pStoreAccountsView.ID).
			FirstOrCreate(&rpStoreAccountsView, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-store-accounts-view"}, RoleID: storeRole.ID, PermissionID: pStoreAccountsView.ID}).Error
	}
	var pStoreAccountsManage model.Permission
	_ = db.Where("name = ?", "store:accounts:manage").First(&pStoreAccountsManage).Error
	if storeRole.ID > 0 && pStoreAccountsManage.ID > 0 {
		var rpStoreAccountsManage model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", storeRole.ID, pStoreAccountsManage.ID).
			FirstOrCreate(&rpStoreAccountsManage, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-store-accounts-manage"}, RoleID: storeRole.ID, PermissionID: pStoreAccountsManage.ID}).Error
	}

	// Grant store withdraw view/apply to store role
	var pStoreWithdrawView model.Permission
	_ = db.Where("name = ?", "store:withdraw:view").First(&pStoreWithdrawView).Error
	if storeRole.ID > 0 && pStoreWithdrawView.ID > 0 {
		var rpStoreWithdrawView model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", storeRole.ID, pStoreWithdrawView.ID).
			FirstOrCreate(&rpStoreWithdrawView, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-store-withdraw-view"}, RoleID: storeRole.ID, PermissionID: pStoreWithdrawView.ID}).Error
	}
	var pStoreWithdrawApply model.Permission
	_ = db.Where("name = ?", "store:withdraw:apply").First(&pStoreWithdrawApply).Error
	if storeRole.ID > 0 && pStoreWithdrawApply.ID > 0 {
		var rpStoreWithdrawApply model.RolePermission
		_ = db.Where("role_id = ? AND permission_id = ?", storeRole.ID, pStoreWithdrawApply.ID).
			FirstOrCreate(&rpStoreWithdrawApply, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-store-withdraw-apply"}, RoleID: storeRole.ID, PermissionID: pStoreWithdrawApply.ID}).Error
	}

		// Grant store coupons view/manage to store role
		var pStoreCouponsView model.Permission
		_ = db.Where("name = ?", "store:coupons:view").First(&pStoreCouponsView).Error
		if storeRole.ID > 0 && pStoreCouponsView.ID > 0 {
			var rpStoreCouponsView model.RolePermission
			_ = db.Where("role_id = ? AND permission_id = ?", storeRole.ID, pStoreCouponsView.ID).
				FirstOrCreate(&rpStoreCouponsView, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-store-coupons-view"}, RoleID: storeRole.ID, PermissionID: pStoreCouponsView.ID}).Error
		}
		var pStoreCouponsManage model.Permission
		_ = db.Where("name = ?", "store:coupons:manage").First(&pStoreCouponsManage).Error
		if storeRole.ID > 0 && pStoreCouponsManage.ID > 0 {
			var rpStoreCouponsManage model.RolePermission
			_ = db.Where("role_id = ? AND permission_id = ?", storeRole.ID, pStoreCouponsManage.ID).
				FirstOrCreate(&rpStoreCouponsManage, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-store-coupons-manage"}, RoleID: storeRole.ID, PermissionID: pStoreCouponsManage.ID}).Error
		}

		// Grant store activities view/manage to store role
		var pStoreActivitiesView model.Permission
		_ = db.Where("name = ?", "store:activities:view").First(&pStoreActivitiesView).Error
		if storeRole.ID > 0 && pStoreActivitiesView.ID > 0 {
			var rpStoreActivitiesView model.RolePermission
			_ = db.Where("role_id = ? AND permission_id = ?", storeRole.ID, pStoreActivitiesView.ID).
				FirstOrCreate(&rpStoreActivitiesView, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-store-activities-view"}, RoleID: storeRole.ID, PermissionID: pStoreActivitiesView.ID}).Error
		}
		var pStoreActivitiesManage model.Permission
		_ = db.Where("name = ?", "store:activities:manage").First(&pStoreActivitiesManage).Error
		if storeRole.ID > 0 && pStoreActivitiesManage.ID > 0 {
			var rpStoreActivitiesManage model.RolePermission
			_ = db.Where("role_id = ? AND permission_id = ?", storeRole.ID, pStoreActivitiesManage.ID).
				FirstOrCreate(&rpStoreActivitiesManage, &model.RolePermission{BaseModel: model.BaseModel{UID: "rp-store-activities-manage"}, RoleID: storeRole.ID, PermissionID: pStoreActivitiesManage.ID}).Error
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
