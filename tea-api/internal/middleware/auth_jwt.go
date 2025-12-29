package middleware

import (
	"net/http"
	"strings"

	"tea-api/internal/pkg/jwtcfg"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthJWT validates Authorization: Bearer <token> and sets user_id in context.
func AuthJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 4010, "message": "缺少或非法的授权头"})
			return
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		cfg := jwtcfg.Get()
		secret := cfg.Secret
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 4011, "message": "令牌无效或已过期"})
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Normalize user_id to uint to align with handlers
			if fuid, ok2 := claims["user_id"].(float64); ok2 {
				if fuid >= 0 {
					c.Set("user_id", uint(fuid))
				}
			} else if iuid, ok2 := claims["user_id"].(int64); ok2 {
				if iuid >= 0 {
					c.Set("user_id", uint(iuid))
				}
			} else if uidStr, ok2 := claims["user_id"].(string); ok2 {
				// attempt to parse numeric string
				var parsed uint
				// simple parse without strconv to minimize deps
				// fallback: do not set if parsing fails
				for i := 0; i < len(uidStr); i++ {
					ch := uidStr[i]
					if ch < '0' || ch > '9' {
						parsed = 0
						// non-numeric; skip setting
						goto NEXT
					}
					parsed = parsed*10 + uint(ch-'0')
				}
				c.Set("user_id", parsed)
			}
			// also expose optional open_id and role for permission checks
			if oid, ok2 := claims["open_id"].(string); ok2 {
				c.Set("open_id", oid)
			}
			if role, ok2 := claims["role"].(string); ok2 {
				c.Set("role", role)
			}
		}

		// 黑/白名单与停用状态拦截（白名单可豁免）
		if v, ok := c.Get("user_id"); ok {
			if uid, ok2 := v.(uint); ok2 {
				if blocked, msg := IsUserBlocked(uid); blocked {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 4031, "message": msg})
					return
				}
			}
		}
	NEXT:
		c.Next()
	}
}
