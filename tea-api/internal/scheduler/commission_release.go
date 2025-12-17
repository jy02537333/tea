package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"tea-api/internal/config"
	"tea-api/internal/service/commission"
	"tea-api/pkg/database"
)

// StartCommissionReleaseScheduler 启动每日佣金解冻调度
func StartCommissionReleaseScheduler() {
	cfg := config.Config.Finance.CommissionRelease
	if !cfg.Enabled {
		zap.L().Info("commission release scheduler disabled")
		return
	}
	// 默认每天 03:00 执行
	hhmm := strings.TrimSpace(cfg.Time)
	if hhmm == "" {
		hhmm = "03:00"
	}
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	loc := time.Local
	if cfg.Timezone != "" {
		if l, err := time.LoadLocation(cfg.Timezone); err == nil {
			loc = l
		}
	}
	go loopDailyCommission(hhmm, batchSize, cfg, loc)
}

func loopDailyCommission(hhmm string, batchSize int, cfg config.CommissionRelease, loc *time.Location) {
	for {
		now := time.Now().In(loc)
		next := nextTimeCommission(now, hhmm, loc)
		wait := time.Until(next)
		zap.L().Info("commission release scheduled", zap.Time("next", next), zap.Duration("wait", wait))
		timer := time.NewTimer(wait)
		<-timer.C
		runCommissionReleaseOnce(next, batchSize, cfg)
	}
}

func nextTimeCommission(now time.Time, hhmm string, loc *time.Location) time.Time {
	parts := strings.Split(hhmm, ":")
	h, m := 3, 0
	if len(parts) == 2 {
		fmt.Sscanf(hhmm, "%d:%d", &h, &m)
	}
	y, mo, d := now.Date()
	t := time.Date(y, mo, d, h, m, 0, 0, loc)
	if !t.After(now) {
		t = t.Add(24 * time.Hour)
	}
	return t
}

func runCommissionReleaseOnce(date time.Time, cfgBatch int, cfg config.CommissionRelease) {
	// Redis 分布式锁（可选）
	if cfg.UseRedisLock {
		if r := database.GetRedis(); r != nil {
			if !acquireCommissionLock(r, date, cfg.LockTTLSecond) {
				zap.L().Info("commission release skipped: lock exists", zap.Time("date", date))
				return
			}
			defer releaseCommissionLock(r, date)
		} else {
			zap.L().Warn("redis not available for commission release, proceeding without distributed lock")
		}
	}

	n, err := commission.ReleaseFrozenCommissions(cfgBatch)
	if err != nil {
		zap.L().Error("commission release run failed", zap.Error(err))
		return
	}
	zap.L().Info("commission release run ok", zap.Int("released", n), zap.Time("date", date))
}

func acquireCommissionLock(r *redis.Client, date time.Time, ttlSec int) bool {
	key := fmt.Sprintf("commission_release:lock:%s", date.Format("2006-01-02"))
	if ttlSec <= 0 {
		ttlSec = 3600
	}
	ok, _ := r.SetNX(context.Background(), key, "1", time.Duration(ttlSec)*time.Second).Result()
	return ok
}

func releaseCommissionLock(r *redis.Client, date time.Time) {
	key := fmt.Sprintf("commission_release:lock:%s", date.Format("2006-01-02"))
	_ = r.Del(context.Background(), key).Err()
}
