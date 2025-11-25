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
	"tea-api/pkg/utils"
)

var configFile = flag.String("config", "configs/config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	// 加载配置
	if err := config.LoadConfig(*configFile); err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	fmt.Println("茶心阁小程序API服务启动成功!")
	fmt.Printf("服务运行在: %s\n", config.Config.Server.Port)

	// 启动HTTP服务器
	startServer()
}

func startServer() {
	// 设置Gin模式
	gin.SetMode(config.Config.Server.Mode)

	// 创建路由
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API路由组
	api := r.Group("/api/v1")

	// 健康检查
	api.GET("/health", func(c *gin.Context) {
		utils.Success(c, gin.H{
			"status":    "ok",
			"message":   "茶心阁小程序API服务正常运行",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	})

	// 模拟用户登录
	api.POST("/user/login", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.InvalidParam(c, err.Error())
			return
		}

		// 模拟登录成功
		mockUser := gin.H{
			"id":       1,
			"uid":      utils.GenerateUID(),
			"open_id":  "mock_openid_123",
			"nickname": "测试用户",
			"avatar":   "",
			"phone":    "",
			"gender":   0,
			"balance":  0.0,
			"points":   0,
		}

		// 生成模拟token
		token := "mock_jwt_token_" + utils.GenerateRandomString(32)

		utils.Success(c, gin.H{
			"token":     token,
			"user_info": mockUser,
		})
	})

	// 获取用户信息
	api.GET("/user/info", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "请先登录")
			return
		}

		// 模拟用户信息
		mockUser := gin.H{
			"id":       1,
			"uid":      "mock_uid_123",
			"open_id":  "mock_openid_123",
			"nickname": "测试用户",
			"avatar":   "",
			"phone":    "",
			"gender":   0,
			"balance":  100.50,
			"points":   50,
		}

		utils.Success(c, mockUser)
	})

	// 商品列表
	api.GET("/products", func(c *gin.Context) {
		mockProducts := []gin.H{
			{
				"id":          1,
				"name":        "西湖龙井",
				"description": "正宗西湖龙井，清香淡雅",
				"price":       168.00,
				"image":       "https://example.com/longjing.jpg",
				"category":    "茶叶",
			},
			{
				"id":          2,
				"name":        "铁观音",
				"description": "安溪铁观音，韵味悠长",
				"price":       128.00,
				"image":       "https://example.com/tieguanyin.jpg",
				"category":    "茶叶",
			},
			{
				"id":          3,
				"name":        "奶茶",
				"description": "香浓奶茶",
				"price":       15.00,
				"image":       "https://example.com/milktea.jpg",
				"category":    "茶饮",
			},
		}

		utils.PageSuccess(c, mockProducts, int64(len(mockProducts)), 1, 10)
	})

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
