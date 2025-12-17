package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	// 统一读取新版 JWT 配置，避免旧版与新版 secret 不一致
	"tea-api/internal/pkg/jwtcfg"
	"tea-api/pkg/utils"
)

// AuthMiddleware JWT认证中间件（旧壳新心）
// 保持旧路由可用，但校验统一走新版 jwtcfg 的 Secret，减少 Token 无效风险。
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.Unauthorized(c, "Token格式错误")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		cfg := jwtcfg.Get()
		secret := cfg.Secret

		// 与新版 AuthJWT 对齐的解析逻辑
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			// 旧版错误码语义保持，但具体校验统一
			utils.Unauthorized(c, "Token无效")
			c.Abort()
			return
		}

		// 兼容旧版上下文注入：尽可能解析 user_id 与 open_id
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// user_id
			if fuid, ok2 := claims["user_id"].(float64); ok2 {
				if fuid >= 0 {
					c.Set("user_id", uint(fuid))
				}
			} else if iuid, ok2 := claims["user_id"].(int64); ok2 {
				if iuid >= 0 {
					c.Set("user_id", uint(iuid))
				}
			} else if uidStr, ok2 := claims["user_id"].(string); ok2 {
				var parsed uint
				for i := 0; i < len(uidStr); i++ {
					ch := uidStr[i]
					if ch < '0' || ch > '9' {
						parsed = 0
						goto NEXTUID
					}
					parsed = parsed*10 + uint(ch-'0')
				}
				c.Set("user_id", parsed)
			}
		NEXTUID:
			// open_id（若存在）
			if oid, ok2 := claims["open_id"].(string); ok2 {
				c.Set("open_id", oid)
			}
			// role（若存在）
			if role, ok2 := claims["role"].(string); ok2 {
				c.Set("role", role)
			}
		}

		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证中间件（与新版 Secret 对齐）
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.Next()
			return
		}

		tokenStr := parts[1]
		cfg := jwtcfg.Get()
		secret := cfg.Secret

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			// 可选认证：解析失败直接放行
			c.Next()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if fuid, ok2 := claims["user_id"].(float64); ok2 {
				if fuid >= 0 {
					c.Set("user_id", uint(fuid))
				}
			} else if iuid, ok2 := claims["user_id"].(int64); ok2 {
				if iuid >= 0 {
					c.Set("user_id", uint(iuid))
				}
			} else if uidStr, ok2 := claims["user_id"].(string); ok2 {
				var parsed uint
				for i := 0; i < len(uidStr); i++ {
					ch := uidStr[i]
					if ch < '0' || ch > '9' {
						parsed = 0
						goto NEXT2
					}
					parsed = parsed*10 + uint(ch-'0')
				}
				c.Set("user_id", parsed)
			}
		NEXT2:
			if oid, ok2 := claims["open_id"].(string); ok2 {
				c.Set("open_id", oid)
			}
		}

		c.Next()
	}
}
