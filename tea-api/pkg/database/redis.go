package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"tea-api/internal/config"
)

var RDB *redis.Client

// InitRedis 初始化Redis连接
func InitRedis() {
	cfg := config.Config.Redis

	fmt.Printf("正在连接Redis: %s\n", cfg.Addr())

	RDB = redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	// 测试连接
	ctx := context.Background()
	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Redis连接失败，但继续启动服务器: %v\n", err)
		fmt.Println("注意：Redis功能将不可用，请确保Redis服务已启动")
		RDB = nil
		return
		// 如果需要强制要求Redis，可以取消下面这行的注释
		// panic(fmt.Errorf("failed to connect redis: %w", err))
	}
}

// GetRedis 获取Redis客户端
func GetRedis() *redis.Client {
	return RDB
}
