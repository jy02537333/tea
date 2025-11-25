//go:build demo
// +build demo

// 演示入口（需 -tags demo 才会参与构建）
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"tea-api/internal/config"
	"tea-api/internal/router"
	"tea-api/pkg/database"
)

var configFile = flag.String("config", "configs/config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	// 加载配置
	if err := config.LoadConfig(*configFile); err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化数据库
	database.InitMySQL()
	database.InitRedis()
	database.InitRabbitMQ()

	fmt.Println("茶心阁小程序API服务启动成功!")
	fmt.Printf("服务运行在: %s\n", config.Config.Server.Port)

	// 启动HTTP服务器
	startServer()
}

func startServer() {
	// 设置Gin模式
	gin.SetMode(config.Config.Server.Mode)

	// 初始化路由
	r := router.SetupRouter()

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         config.Config.Server.Port,
		Handler:      r,
		ReadTimeout:  config.Config.Server.ReadTimeout * time.Second,
		WriteTimeout: config.Config.Server.WriteTimeout * time.Second,
	}

	// 启动服务器
	fmt.Printf("服务器启动在端口: %s\n", config.Config.Server.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
