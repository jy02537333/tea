package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

// OperationLogMiddleware 记录管理员变更操作日志（仅对 /api/v1/admin/* 的 POST/PUT/DELETE 生效）
func OperationLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := strings.ToUpper(c.Request.Method)
		// 读取配置（默认启用）
		cfg := config.Config.Observability.OperationLog
		if !cfg.Enabled {
			c.Next()
			return
		}
		// 基础条件：Admin 路径 + 变更方法
		if !strings.HasPrefix(path, "/api/v1/admin/") || (method != "POST" && method != "PUT" && method != "DELETE") {
			c.Next()
			return
		}
		// include 前缀（若配置则要求命中其中之一）
		if len(cfg.IncludePrefixes) > 0 {
			matched := false
			for _, p := range cfg.IncludePrefixes {
				if p != "" && strings.HasPrefix(path, p) {
					matched = true
					break
				}
			}
			if !matched {
				c.Next()
				return
			}
		}
		// exclude 前缀（若配置则命中则跳过记录）
		if len(cfg.ExcludePrefixes) > 0 {
			for _, p := range cfg.ExcludePrefixes {
				if p != "" && strings.HasPrefix(path, p) {
					c.Next()
					return
				}
			}
		}

		// 读取请求体（最多 20KB）并回填
		var body []byte
		if c.Request.Body != nil {
			lr := io.LimitReader(c.Request.Body, 20*1024)
			body, _ = io.ReadAll(lr)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		start := time.Now()
		c.Next()
		_ = start // 预留耗时记录需要

		// 仅在成功（2xx/3xx）时记录
		status := c.Writer.Status()
		if status < 200 || status >= 400 {
			return
		}

		// 取用户ID
		var userID uint
		if v, ok := c.Get("user_id"); ok {
			if id, ok2 := v.(uint); ok2 {
				userID = id
			}
		}

		// 模块判定
		module := "admin"
		if strings.Contains(path, "/rbac/") {
			module = "rbac"
		} else if strings.Contains(path, "/accrual/") {
			module = "finance"
		}

		// 记录
		db := database.GetDB()
		if db != nil {
			_ = db.Create(&model.OperationLog{
				BaseModel:   model.BaseModel{UID: utils.GenerateUID()},
				UserID:      userID,
				Module:      module,
				Operation:   method + " " + path,
				Description: "",
				RequestData: string(body),
				IP:          c.ClientIP(),
				UserAgent:   c.Request.UserAgent(),
			}).Error
		}
	}
}
