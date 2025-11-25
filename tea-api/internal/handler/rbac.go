package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

type RBACHandler struct{}

func NewRBACHandler() *RBACHandler { return &RBACHandler{} }

// GET /api/v1/admin/rbac/roles
func (h *RBACHandler) ListRoles(c *gin.Context) {
	var list []model.Role
	if err := database.GetDB().Order("id asc").Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, list)
}

// GET /api/v1/admin/rbac/permissions
func (h *RBACHandler) ListPermissions(c *gin.Context) {
	var list []model.Permission
	if err := database.GetDB().Order("id asc").Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, list)
}

// GET /api/v1/admin/rbac/role-permissions?role_id=1
func (h *RBACHandler) ListRolePermissions(c *gin.Context) {
	rid, _ := strconv.Atoi(c.DefaultQuery("role_id", "0"))
	if rid <= 0 {
		utils.InvalidParam(c, "role_id required")
		return
	}
	var list []model.RolePermission
	if err := database.GetDB().Where("role_id = ?", rid).Order("id asc").Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, list)
}

// GET /api/v1/admin/rbac/user-permissions?user_id=1
func (h *RBACHandler) ListUserPermissions(c *gin.Context) {
	uid, _ := strconv.Atoi(c.DefaultQuery("user_id", "0"))
	if uid <= 0 {
		utils.InvalidParam(c, "user_id required")
		return
	}
	db := database.GetDB()
	type row struct{ Name string }
	var names []row
	if err := db.Table("permissions").
		Select("permissions.name as name").
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Joins("JOIN user_roles ur ON ur.role_id = rp.role_id").
		Where("ur.user_id = ?", uid).Scan(&names).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, names)
}

// POST /api/v1/admin/rbac/cache/invalidate  {"user_id":123}  // 若不传 user_id 可扩展为全量清空
func (h *RBACHandler) InvalidateCache(c *gin.Context) {
	var req struct {
		UserID uint `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.InvalidParam(c, err.Error())
		return
	}
	if req.UserID == 0 {
		// 为安全起见，目前仅允许按用户清空，如需全量清空可以扩展额外参数确认
		utils.InvalidParam(c, "user_id required")
		return
	}
	service.InvalidateUserPermCache(req.UserID)
	utils.Success(c, "ok")
}

// POST /api/v1/admin/rbac/role  {"name":"auditor2","display_name":"审计员2"}
func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		utils.InvalidParam(c, "name required")
		return
	}
	role, err := service.CreateRole(database.GetDB(), req.Name, req.DisplayName)
	if err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, role)
}

// DELETE /api/v1/admin/rbac/role/:id
func (h *RBACHandler) DeleteRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		utils.InvalidParam(c, "id invalid")
		return
	}
	if err := database.GetDB().Delete(&model.Role{}, id).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, "ok")
}

// POST /api/v1/admin/rbac/permission {name, display_name, module, action, resource}
func (h *RBACHandler) CreatePermission(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Module      string `json:"module"`
		Action      string `json:"action"`
		Resource    string `json:"resource"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		utils.InvalidParam(c, "name required")
		return
	}
	perm, err := service.CreatePermission(database.GetDB(), req.Name, req.DisplayName, req.Module, req.Action, req.Resource)
	if err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, perm)
}

// POST /api/v1/admin/rbac/role/assign-permission {role_id, permission_id}
func (h *RBACHandler) AssignPermissionToRole(c *gin.Context) {
	var req struct {
		RoleID       uint `json:"role_id"`
		PermissionID uint `json:"permission_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RoleID == 0 || req.PermissionID == 0 {
		utils.InvalidParam(c, "role_id & permission_id required")
		return
	}
	if err := service.AssignPermissionToRole(database.GetDB(), req.RoleID, req.PermissionID); err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, "ok")
}

// POST /api/v1/admin/rbac/role/revoke-permission {role_id, permission_id}
func (h *RBACHandler) RevokePermissionFromRole(c *gin.Context) {
	var req struct {
		RoleID       uint `json:"role_id"`
		PermissionID uint `json:"permission_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RoleID == 0 || req.PermissionID == 0 {
		utils.InvalidParam(c, "role_id & permission_id required")
		return
	}
	if err := service.RevokePermissionFromRole(database.GetDB(), req.RoleID, req.PermissionID); err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, "ok")
}

// POST /api/v1/admin/rbac/role/assign-permissions {role_id, permission_ids: [1,2,3]}
func (h *RBACHandler) AssignPermissionsToRole(c *gin.Context) {
	var req struct {
		RoleID        uint   `json:"role_id"`
		PermissionIDs []uint `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RoleID == 0 || len(req.PermissionIDs) == 0 {
		utils.InvalidParam(c, "role_id & permission_ids required")
		return
	}
	if err := service.AssignPermissionsToRole(database.GetDB(), req.RoleID, req.PermissionIDs); err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, "ok")
}

// POST /api/v1/admin/rbac/user/assign-role {user_id, role_id}
func (h *RBACHandler) AssignRoleToUser(c *gin.Context) {
	var req struct {
		UserID uint `json:"user_id"`
		RoleID uint `json:"role_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == 0 || req.RoleID == 0 {
		utils.InvalidParam(c, "user_id & role_id required")
		return
	}
	if err := service.AssignRoleToUser(database.GetDB(), req.UserID, req.RoleID); err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, "ok")
}

// POST /api/v1/admin/rbac/user/revoke-role {user_id, role_id}
func (h *RBACHandler) RevokeRoleFromUser(c *gin.Context) {
	var req struct {
		UserID uint `json:"user_id"`
		RoleID uint `json:"role_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == 0 || req.RoleID == 0 {
		utils.InvalidParam(c, "user_id & role_id required")
		return
	}
	if err := service.RevokeRoleFromUser(database.GetDB(), req.UserID, req.RoleID); err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, "ok")
}
