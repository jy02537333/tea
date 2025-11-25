package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 简单的基于 IP 的滑动窗口限流：每个 IP 每分钟最多允许 N 次请求
const defaultModelLimitPerMinute = 30

type ipQuota struct {
	Count       int
	WindowStart time.Time
}

var (
	quotaMu sync.Mutex
	quotas  = make(map[string]*ipQuota)
)

// ModelRateLimit 返回一个中间件，限制每个 IP 每分钟请求次数
func ModelRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只限模型相关的 POST 请求（更严格的路由级别也会应用）
		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		ip := c.ClientIP()
		now := time.Now()

		quotaMu.Lock()
		q, ok := quotas[ip]
		if !ok || now.Sub(q.WindowStart) > time.Minute {
			quotas[ip] = &ipQuota{Count: 1, WindowStart: now}
			quotaMu.Unlock()
			c.Next()
			return
		}

		if q.Count >= defaultModelLimitPerMinute {
			quotaMu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "模型调用过于频繁，请稍后重试",
			})
			return
		}

		q.Count++
		quotaMu.Unlock()
		c.Next()
	}
}
