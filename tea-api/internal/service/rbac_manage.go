package service

import (
	"fmt"

	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

func dbOrDefault(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return database.GetDB()
}

// CreateRole 创建角色（若存在则返回已存在的记录）
func CreateRole(db *gorm.DB, name, displayName string) (*model.Role, error) {
	d := dbOrDefault(db)
	if d == nil {
		return nil, fmt.Errorf("db is nil")
	}
	role := &model.Role{BaseModel: model.BaseModel{UID: "role-" + name}, Name: name, DisplayName: displayName}
	if err := d.Where("name = ?", name).FirstOrCreate(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

// CreatePermission 创建权限（若存在则返回已存在的记录）
func CreatePermission(db *gorm.DB, name, displayName, module, action, resource string) (*model.Permission, error) {
	d := dbOrDefault(db)
	if d == nil {
		return nil, fmt.Errorf("db is nil")
	}
	perm := &model.Permission{BaseModel: model.BaseModel{UID: "perm-" + name}, Name: name, DisplayName: displayName, Module: module, Action: action, Resource: resource}
	if err := d.Where("name = ?", name).FirstOrCreate(perm).Error; err != nil {
		return nil, err
	}
	return perm, nil
}

// AssignPermissionToRole 赋予角色一个权限，并失效该角色所有用户的权限缓存
func AssignPermissionToRole(db *gorm.DB, roleID, permID uint) error {
	d := dbOrDefault(db)
	if d == nil {
		return fmt.Errorf("db is nil")
	}
	var rp model.RolePermission
	if err := d.Where("role_id = ? AND permission_id = ?", roleID, permID).
		FirstOrCreate(&rp, &model.RolePermission{BaseModel: model.BaseModel{UID: fmt.Sprintf("rp-%d-%d", roleID, permID)}, RoleID: roleID, PermissionID: permID}).Error; err != nil {
		return err
	}
	invalidateUsersByRole(d, roleID)
	return nil
}

// RevokePermissionFromRole 从角色移除一个权限，并失效该角色所有用户的权限缓存
func RevokePermissionFromRole(db *gorm.DB, roleID, permID uint) error {
	d := dbOrDefault(db)
	if d == nil {
		return fmt.Errorf("db is nil")
	}
	if err := d.Where("role_id = ? AND permission_id = ?", roleID, permID).Delete(&model.RolePermission{}).Error; err != nil {
		return err
	}
	invalidateUsersByRole(d, roleID)
	return nil
}

// AssignPermissionsToRole 批量赋予权限给角色（幂等），并统一失效该角色下用户缓存
func AssignPermissionsToRole(db *gorm.DB, roleID uint, permIDs []uint) error {
	d := dbOrDefault(db)
	if d == nil {
		return fmt.Errorf("db is nil")
	}
	if roleID == 0 || len(permIDs) == 0 {
		return nil
	}
	for _, pid := range permIDs {
		var rp model.RolePermission
		if err := d.Where("role_id = ? AND permission_id = ?", roleID, pid).
			FirstOrCreate(&rp, &model.RolePermission{BaseModel: model.BaseModel{UID: fmt.Sprintf("rp-%d-%d", roleID, pid)}, RoleID: roleID, PermissionID: pid}).Error; err != nil {
			return err
		}
	}
	invalidateUsersByRole(d, roleID)
	return nil
}

// AssignRoleToUser 给用户赋予角色，并失效该用户的权限缓存
func AssignRoleToUser(db *gorm.DB, userID, roleID uint) error {
	d := dbOrDefault(db)
	if d == nil {
		return fmt.Errorf("db is nil")
	}
	var ur model.UserRole
	if err := d.Where("user_id = ? AND role_id = ?", userID, roleID).
		FirstOrCreate(&ur, &model.UserRole{BaseModel: model.BaseModel{UID: fmt.Sprintf("ur-%d-%d", userID, roleID)}, UserID: userID, RoleID: roleID}).Error; err != nil {
		return err
	}
	InvalidateUserPermCache(userID)
	return nil
}

// RevokeRoleFromUser 移除用户的某个角色，并失效该用户的权限缓存
func RevokeRoleFromUser(db *gorm.DB, userID, roleID uint) error {
	d := dbOrDefault(db)
	if d == nil {
		return fmt.Errorf("db is nil")
	}
	if err := d.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&model.UserRole{}).Error; err != nil {
		return err
	}
	InvalidateUserPermCache(userID)
	return nil
}

// invalidateUsersByRole 查找拥有该角色的用户并逐个失效权限缓存
func invalidateUsersByRole(db *gorm.DB, roleID uint) {
	var uids []uint
	_ = db.Model(&model.UserRole{}).Select("user_id").Where("role_id = ?", roleID).Scan(&uids).Error
	for _, uid := range uids {
		InvalidateUserPermCache(uid)
	}
}
