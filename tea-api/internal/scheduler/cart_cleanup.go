package scheduler

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"tea-api/internal/service"
	"tea-api/pkg/database"
)

// StartCartCleanupScheduler 启动购物车超时清理：超过 30 分钟的购物车条目将被删除，并回补其占用库存。
// 可通过环境变量控制：
// - TEA_CART_CLEANUP_ENABLED=0 关闭（默认开启）
// - TEA_CART_CLEANUP_MAX_AGE_MIN=30 过期分钟数
// - TEA_CART_CLEANUP_INTERVAL_MIN=5 扫描间隔分钟数
// - TEA_CART_CLEANUP_BATCH=200 每次最多处理条数
// - TEA_CART_CLEANUP_USE_REDIS_LOCK=1 使用 Redis 分布式锁（默认关闭）
func StartCartCleanupScheduler() {
	if os.Getenv("TEA_CART_CLEANUP_ENABLED") == "0" {
		zap.L().Info("cart cleanup scheduler disabled")
		return
	}
	maxAgeMin := envInt("TEA_CART_CLEANUP_MAX_AGE_MIN", 30)
	intervalMin := envInt("TEA_CART_CLEANUP_INTERVAL_MIN", 5)
	batch := envInt("TEA_CART_CLEANUP_BATCH", 200)
	useRedisLock := os.Getenv("TEA_CART_CLEANUP_USE_REDIS_LOCK") == "1"

	maxAge := time.Duration(maxAgeMin) * time.Minute
	interval := time.Duration(intervalMin) * time.Minute
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	if batch <= 0 {
		batch = 200
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		zap.L().Info("cart cleanup scheduler started",
			zap.Duration("max_age", maxAge),
			zap.Duration("interval", interval),
			zap.Int("batch", batch),
			zap.Bool("redis_lock", useRedisLock),
		)
		for {
			<-ticker.C
			releaseLock := func() {}
			if useRedisLock {
				r := database.GetRedis()
				if r == nil {
					zap.L().Warn("cart cleanup: redis not available, proceeding without lock")
				} else {
					ok, err := acquireCartCleanupLock(r, 300)
					if err != nil {
						zap.L().Warn("cart cleanup: acquire lock failed", zap.Error(err))
						continue
					}
					if !ok {
						continue
					}
					releaseLock = func() { releaseCartCleanupLock(r) }
				}
			}

			svc := service.NewCartService()
			n, err := svc.CleanupExpiredItems(maxAge, batch)
			if err != nil {
				zap.L().Error("cart cleanup failed", zap.Error(err))
				releaseLock()
				continue
			}
			if n > 0 {
				zap.L().Info("cart cleanup ok", zap.Int("deleted", n))
			}
			releaseLock()
		}
	}()
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func acquireCartCleanupLock(r *redis.Client, ttlSec int) (bool, error) {
	if ttlSec <= 0 {
		ttlSec = 300
	}
	key := "cart_cleanup:lock"
	ok, err := r.SetNX(context.Background(), key, fmt.Sprintf("%d", time.Now().Unix()), time.Duration(ttlSec)*time.Second).Result()
	return ok, err
}

func releaseCartCleanupLock(r *redis.Client) {
	_ = r.Del(context.Background(), "cart_cleanup:lock").Err()
}
