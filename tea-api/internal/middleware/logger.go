package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/logx"
	"tea-api/pkg/utils"
)

// AccessLogMiddleware 访问日志中间件
func AccessLogMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 记录访问日志到数据库
		go func() {
			var userID *uint
			if uid, exists := param.Keys["user_id"]; exists {
				if u, ok := uid.(uint); ok {
					userID = makeUintPtr(u)
				}
			}

			accessLog := &model.AccessLog{
				BaseModel: model.BaseModel{
					UID: utils.GenerateUID(),
				},
				UserID:     userID,
				Method:     param.Method,
				Path:       param.Path,
				Query:      param.Request.URL.RawQuery,
				UserAgent:  param.Request.UserAgent(),
				IP:         param.ClientIP,
				StatusCode: param.StatusCode,
				Latency:    int64(param.Latency),
			}

			// 如果是POST/PUT/PATCH请求，记录请求体
			if param.Method == "POST" || param.Method == "PUT" || param.Method == "PATCH" {
				// 注意：这里需要从gin.Context中获取请求体，但在日志格式化器中无法直接获取
				// 实际项目中建议使用专门的中间件来处理
			}

			db := database.GetDB()
			if err := db.Create(accessLog).Error; err != nil {
				// 记录错误日志
				zap.L().Error("Failed to create access log", zap.Error(err))
			}
		}()

		// 返回空字符串，不在控制台输出
		return ""
	})
}

// DetailedAccessLogMiddleware 详细访问日志中间件
func DetailedAccessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录详细访问日志
		go func() {
			var userID *uint
			if uid, exists := c.Get("user_id"); exists {
				if u, ok := uid.(uint); ok {
					userID = makeUintPtr(u)
				}
			}
			var role string
			if r, ok := c.Get("role"); ok {
				if s, ok := r.(string); ok {
					role = s
				}
			}
			var rid string
			if v, ok := c.Get("request_id"); ok {
				if s, ok := v.(string); ok {
					rid = s
				}
			}

			duration := time.Since(start)

			accessLog := &model.AccessLog{
				BaseModel: model.BaseModel{
					UID: utils.GenerateUID(),
				},
				UserID:     userID,
				Method:     c.Request.Method,
				Path:       c.Request.URL.Path,
				Query:      c.Request.URL.RawQuery,
				UserAgent:  c.Request.UserAgent(),
				IP:         c.ClientIP(),
				StatusCode: c.Writer.Status(),
				Latency:    duration.Nanoseconds(),
			}

			db := database.GetDB()
			if db != nil {
				if err := db.Create(accessLog).Error; err != nil {
					// 若因缺表失败，尝试迁移并重试一次
					msg := err.Error()
					if msg == "no such table: access_logs" || contains(msg, "Error 1146") {
						_ = db.Migrator().AutoMigrate(&model.AccessLog{})
						_ = db.Create(accessLog).Error
					} else {
						zap.L().Error("Failed to create detailed access log", zap.Error(err))
					}
				}
			}
			// 结构化日志输出到 Zap
			logUserID := uint(0)
			if userID != nil {
				logUserID = *userID
			}
			logx.Get().Info("access",
				zap.String("rid", rid),
				zap.String("method", accessLog.Method),
				zap.String("path", accessLog.Path),
				zap.String("query", accessLog.Query),
				zap.Int("status", accessLog.StatusCode),
				zap.Duration("latency", duration),
				zap.String("ip", accessLog.IP),
				zap.String("ua", accessLog.UserAgent),
				zap.Uint("user_id", logUserID),
				zap.String("role", role),
			)
		}()
	}
}

func makeUintPtr(val uint) *uint {
	if val == 0 {
		return nil
	}
	copy := val
	return &copy
}

// contains reports whether substr is within s.
func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && (indexOf(s, substr) >= 0))
}

// indexOf returns the index of substr in s, or -1 if not found.
func indexOf(s, substr string) int {
	// 简化实现：使用标准库 strings（避免引入依赖）
	// 保持文件内自足性
	// 注意：Go 编译器内联优化足够，这里直接调用 strings.Index
	return strings.Index(s, substr)
}
