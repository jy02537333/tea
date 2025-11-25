package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"tea-api/pkg/utils"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// 检查token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.Unauthorized(c, "Token格式错误")
			c.Abort()
			return
		}

		// 解析token
		token := parts[1]
		claims, err := utils.ParseToken(token)
		if err != nil {
			utils.Unauthorized(c, "Token无效")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("open_id", claims.OpenID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证中间件
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// 检查token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.Next()
			return
		}

		// 解析token
		token := parts[1]
		claims, err := utils.ParseToken(token)
		if err != nil {
			c.Next()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("open_id", claims.OpenID)

		c.Next()
	}
}
