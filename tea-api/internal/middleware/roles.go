package middleware

import (
	"github.com/gin-gonic/gin"

	"tea-api/pkg/utils"
)

// RequireRoles 基于简化的 User.Role 字段进行角色校验
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先要求已认证
		_, exists := c.Get("user_id")
		if !exists {
			utils.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// 从上下文读取角色，避免额外DB查询
		roleVal, ok := c.Get("role")
		if !ok {
			utils.ServerError(c, "缺少角色信息")
			c.Abort()
			return
		}
		role, _ := roleVal.(string)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}

		utils.Forbidden(c, "权限不足")
		c.Abort()
	}
}
