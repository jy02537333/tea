package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

// RequireAccrualPermission 仅允许具备计息权限的角色访问（向后兼容，内部转到 RequirePermission）
func RequireAccrualPermission() gin.HandlerFunc { return RequirePermission("accrual:run") }

// RequirePermission 校验当前用户是否具备指定权限（DB优先，配置回退，admin 永远放行）
// permName: 形如 "accrual:run"、"accrual:export" 等
func RequirePermission(permName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// admin 角色直通
		if v, ok := c.Get("role"); ok {
			if role, _ := v.(string); strings.EqualFold(role, "admin") {
				c.Next()
				return
			}
		}

		// 尝试DB鉴权：User -> UserRole -> RolePermission -> Permission(name)
		if hasPermDB(c, permName) {
			c.Next()
			return
		}

		// 回退到配置的 allowed_roles（与旧逻辑兼容）
		if v, ok := c.Get("role"); ok {
			if role, _ := v.(string); roleAllowedByConfig(role) {
				c.Next()
				return
			}
		}

		utils.Forbidden(c, "insufficient permission")
		c.Abort()
	}
}

func roleAllowedByConfig(role string) bool {
	if strings.EqualFold(role, "admin") {
		return true
	}
	for _, r := range config.Config.Finance.Accrual.AllowedRoles {
		if strings.EqualFold(r, role) {
			return true
		}
	}
	return false
}

func hasPermDB(c *gin.Context, permName string) bool {
	db := database.GetDB()
	if db == nil {
		return false
	}

	// 获取当前用户ID
	v, ok := c.Get("user_id")
	if !ok {
		return false
	}
	uid, ok := v.(uint)
	if !ok || uid == 0 {
		return false
	}

	// 通过服务获取用户权限（包含 Redis 缓存）。
	// 注意：如果成功拿到列表，则信任该结果，不再回落到 DB 直查，确保“缓存一致性”语义（未失效前不生效）。
	perms, err := service.GetUserPermissions(db, uid)
	if err == nil {
		want := strings.ToLower(permName)
		for _, p := range perms {
			if p == want {
				return true
			}
		}
		// 成功返回但未命中，按未授权处理（不再做 DB 回落）
		return false
	}

	// 当且仅当获取权限列表出错时，兜底：直接用 DB 做一次判断
	var perm model.Permission
	if e := db.Where("name = ?", permName).First(&perm).Error; e != nil {
		if e == gorm.ErrRecordNotFound {
			return false
		}
		return false
	}
	var cnt int64
	e := db.Model(&model.RolePermission{}).
		Joins("JOIN user_roles ur ON ur.role_id = role_permissions.role_id").
		Where("ur.user_id = ? AND role_permissions.permission_id = ?", uid, perm.ID).
		Count(&cnt).Error
	if e != nil {
		return false
	}
	return cnt > 0
}
