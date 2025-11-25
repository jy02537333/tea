package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"tea-api/internal/config"
	"tea-api/pkg/database"
)

var configFile = flag.String("config", "configs/config.yaml", "配置文件路径")

func main() {
	flag.Parse()

	if err := config.LoadConfig(*configFile); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(2)
	}

	// 初始化 Redis
	database.InitRedis()

	r := database.GetRedis()
	if r == nil {
		fmt.Println("Redis 连接失败: client 为 nil (请检查地址/网络/服务是否可用)")
		os.Exit(1)
	}

	if pong, err := r.Ping(context.Background()).Result(); err != nil {
		fmt.Printf("Redis PING 失败: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("Redis 连接成功，PING=%s，地址=%s\n", pong, config.Config.Redis.Addr())
	}
}
