package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	repo "tea-api/internal/repo"
	svc "tea-api/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"tea-api/internal/config"
	"tea-api/internal/pkg/jwtcfg"
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

	applyServerPortOverrideFromEnv()

	// 在初始化数据库之前打印关键运行配置（JWT 与路由占位）
	cfg := jwtcfg.Get()
	secret := cfg.Secret
	src := "file"
	if os.Getenv("TEA_JWT_SECRET") != "" {
		src = "env"
	}
	fmt.Printf("JWT secret_len: %d, source: %s, exp_min: %d\n",
		len(secret), src, cfg.ExpiryMinutes)

	// 初始化数据库（MySQL/禁迁移可切换）
	database.InitDatabase()
	database.InitRedis()

	// 注入用户汇总的依赖（全局）
	svc.SetSummaryDeps(svc.SummaryDeps{
		Wallet:     repo.NewWalletRepository(),
		Points:     repo.NewPointsRepository(),
		Coupons:    repo.NewCouponsRepository(),
		Membership: repo.NewMembershipRepository(),
	})

	// 启动计息调度器（若启用）
	scheduler.StartAccrualScheduler()
	// 启动佣金解冻调度（若启用）
	scheduler.StartCommissionReleaseScheduler()

	fmt.Println("茶心阁小程序API服务初始化完成")
	fmt.Printf("服务准备监听: %s\n", config.Config.Server.Port)

	// 启动HTTP服务器
	startServer()
}

func applyServerPortOverrideFromEnv() {
	// Convenience for local/dev/CI: allow overriding port without editing YAML.
	// Accept values like "9292" or ":9292".
	if raw := strings.TrimSpace(os.Getenv("TEA_SERVER_PORT")); raw != "" {
		p := strings.TrimPrefix(raw, ":")
		if p != "" {
			config.Config.Server.Port = ":" + p
		}
	}
}

func startServer() {
	// 设置Gin模式
	gin.SetMode(config.Config.Server.Mode)

	// 初始化路由
	r := router.SetupRouter()

	// 路由表打印（仅 debug/test 模式）
	if config.Config.Server.Mode != "release" {
		fmt.Println("Registered routes:")
		for _, ri := range r.Routes() {
			fmt.Printf("- %-6s %s -> %s\n", ri.Method, ri.Path, ri.Handler)
		}
	}

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
		if strings.Contains(err.Error(), "address already in use") {
			log.Fatalf("服务器启动失败: %v (提示: 端口被占用，可设置 TEA_SERVER_PORT=9393 或使用 ./run-tea-api.sh 自动释放端口)", err)
		}
		log.Fatalf("服务器启动失败: %v", err)
	}
}
