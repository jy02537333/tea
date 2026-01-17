package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

type storeAdminRow struct {
	StoreID uint `gorm:"column:store_id"`
}

// RequireStoreScope enforces that store admins (role=store) can only access their own store resources.
// It checks store_admins(user_id -> store_id) and requires that c.Param(paramName) equals that store_id.
func RequireStoreScope(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, _ := c.Get("role")
		role, _ := roleVal.(string)
		// Admin bypasses store binding checks.
		if role == "admin" {
			c.Next()
			return
		}

		uidVal, ok := c.Get("user_id")
		if !ok {
			utils.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}
		uid, ok := uidVal.(uint)
		if !ok || uid == 0 {
			utils.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		idStr := c.Param(paramName)
		requestedID64, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || requestedID64 == 0 {
			utils.InvalidParam(c, "缺少有效的门店ID")
			c.Abort()
			return
		}
		requestedID := uint(requestedID64)

		db := database.GetDB()
		if db == nil {
			utils.Forbidden(c, "无法校验门店权限")
			c.Abort()
			return
		}

		var rows []storeAdminRow
		if err := db.Table("store_admins").Select("store_id").Where("user_id = ?", uid).Order("id desc").Limit(1).Find(&rows).Error; err != nil {
			utils.Forbidden(c, "无法校验门店权限")
			c.Abort()
			return
		}
		if len(rows) == 0 || rows[0].StoreID == 0 {
			utils.Forbidden(c, "无门店权限")
			c.Abort()
			return
		}

		if rows[0].StoreID != requestedID {
			utils.Forbidden(c, "门店越权")
			c.Abort()
			return
		}

		c.Set("store_id", rows[0].StoreID)
		c.Next()
	}
}
