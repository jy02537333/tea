package middleware

import (
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestIDMiddleware 注入 X-Request-ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = genID()
		}
		c.Writer.Header().Set("X-Request-ID", rid)
		c.Set("request_id", rid)
		c.Next()
	}
}

func genID() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return time.Now().Format("20060102T150405") + "-" + randomString(r, 8)
}

func randomString(r *rand.Rand, n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
