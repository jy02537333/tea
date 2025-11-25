package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"tea-api/internal/config"
	"tea-api/internal/router"
	"tea-api/internal/scheduler"
	"tea-api/pkg/database"
	"tea-api/pkg/logx"
)

var configFile = flag.String("config", "configs/config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	// 加载配置
	if err := config.LoadConfig(*configFile); err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化日志
	if err := logx.Init(); err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}
	defer logx.Sync()

	// 初始化数据库（MySQL）
	database.InitDatabase()
	database.InitRedis()

	// 启动计息调度器（若启用）
	scheduler.StartAccrualScheduler()

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
		ReadTimeout:  time.Duration(config.Config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.Config.Server.WriteTimeout) * time.Second,
	}

	// 启动服务器
	fmt.Printf("服务器启动在 %s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logx.Get().Error("服务器启动失败", zap.Error(err))
		log.Fatalf("服务器启动失败: %v", err)
	}
}
