package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"tea-api/internal/config"
	"tea-api/internal/service"
	"tea-api/pkg/database"
)

// StartAccrualScheduler 启动每日自动计息调度
func StartAccrualScheduler() {
	cfg := config.Config.Finance.Accrual
	if !cfg.Enabled {
		zap.L().Info("accrual scheduler disabled")
		return
	}
	hhmm := strings.TrimSpace(cfg.Time)
	if hhmm == "" {
		hhmm = "02:00"
	}
	rate := cfg.Rate
	if rate <= 0 {
		rate = 0.001
	}
	loc := time.Local
	if cfg.Timezone != "" {
		if l, err := time.LoadLocation(cfg.Timezone); err == nil {
			loc = l
		}
	}
	go loopDaily(hhmm, rate, cfg, loc)
}

func loopDaily(hhmm string, rate float64, cfg config.Accrual, loc *time.Location) {
	for {
		now := time.Now().In(loc)
		next := nextTime(now, hhmm, loc)
		// 如果配置跳过周末/节假日，则找到下一个工作日
		for shouldSkip(next, cfg) {
			next = next.Add(24 * time.Hour)
			next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), 0, 0, loc)
		}
		wait := time.Until(next)
		zap.L().Info("accrual scheduled", zap.Time("next", next), zap.Duration("wait", wait), zap.Float64("rate", rate))
		timer := time.NewTimer(wait)
		<-timer.C
		runOnce(next, rate, cfg)
	}
}

func nextTime(now time.Time, hhmm string, loc *time.Location) time.Time {
	parts := strings.Split(hhmm, ":")
	h, m := 2, 0
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

func shouldSkip(t time.Time, cfg config.Accrual) bool {
	if cfg.SkipWeekends {
		wd := t.Weekday()
		if wd == time.Saturday || wd == time.Sunday {
			return true
		}
	}
	if len(cfg.Holidays) > 0 {
		d := t.Format("2006-01-02")
		for _, h := range cfg.Holidays {
			if strings.TrimSpace(h) == d {
				return true
			}
		}
	}
	return false
}

func runOnce(date time.Time, rate float64, cfg config.Accrual) {
	// Redis 分布式锁（可选）
	if cfg.UseRedisLock {
		if r := database.GetRedis(); r != nil {
			if !acquireLock(r, date, cfg.LockTTLSecond) {
				zap.L().Info("accrual skipped: lock exists", zap.Time("date", date))
				return
			}
			defer releaseLock(r, date)
		} else {
			zap.L().Warn("redis not available, proceeding without distributed lock")
		}
	}
	svc := service.NewAccrualService()
	n, err := svc.Run(date, rate, nil)
	if err != nil {
		zap.L().Error("accrual run failed", zap.Error(err))
		return
	}
	zap.L().Info("accrual run ok", zap.Int("updated", n), zap.Time("date", date), zap.Float64("rate", rate))
}

func acquireLock(r *redis.Client, date time.Time, ttlSec int) bool {
	key := fmt.Sprintf("accrual:lock:%s", date.Format("2006-01-02"))
	if ttlSec <= 0 {
		ttlSec = 3600
	}
	ok, _ := r.SetNX(context.Background(), key, "1", time.Duration(ttlSec)*time.Second).Result()
	return ok
}

func releaseLock(r *redis.Client, date time.Time) {
	key := fmt.Sprintf("accrual:lock:%s", date.Format("2006-01-02"))
	_ = r.Del(context.Background(), key).Err()
}
