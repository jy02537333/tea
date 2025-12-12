//go:build demo
// +build demo

// 无迁移模式演示入口（需 -tags demo）
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"tea-api/internal/config"
	"tea-api/internal/router"
	"tea-api/pkg/database"

	"github.com/gin-gonic/gin"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	log.Println("🚀 启动茶心阁API服务器（无迁移模式）...")

	// 加载配置
	if err := config.LoadConfig("configs/config.yaml"); err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化数据库连接（不执行迁移）
	_, err := database.InitWithoutMigrate()
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	log.Println("✅ 数据库连接成功（跳过迁移）")

	// 设置Gin模式为调试模式
	gin.SetMode(gin.DebugMode)

	// 使用现有的路由设置
	r := router.SetupRouter()

	// 添加CORS中间件
	r.Use(corsMiddleware())

	// 添加顶层探针，便于快速确认请求是否到达当前 Gin 引擎
	r.GET("/_probe", func(c *gin.Context) {
		c.JSON(200, gin.H{"data": gin.H{"ok": true, "engine": "main_no_migrate"}})
	})

	// 在启动前打印所有已注册路由，帮助定位 404 原因（确认实际生效的路由）
	for _, ri := range r.Routes() {
		log.Printf("route: %-6s %s -> %s", ri.Method, ri.Path, ri.Handler)
	}

	// `router.SetupRouter()` 已包含健康检查端点 `/api/v1/health`，不重复注册以避免冲突。

	// 启动服务器
	port := ":9292"
	log.Printf("🚀 服务器启动在端口 %s", port)
	log.Printf("🔗 健康检查: http://localhost%s/api/v1/health", port)

	// 优雅关闭
	go func() {
		if err := r.Run(port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 服务器正在关闭...")

	log.Println("✅ 服务器已优雅关闭")
}
