package scheduler

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"tea-api/internal/config"
	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

// 验证 runOnce 的 Redis 分布式锁（若 Redis 不可用则跳过）
func Test_Accrual_RedisLock_Skips_Duplicate(t *testing.T) {
	// 如果希望强制使用 MySQL，确保 TEA_USE_SQLITE=0；否则保留原有 sqlite 路径
	_ = os.Setenv("TEA_USE_SQLITE", "0")
	_ = os.Setenv("TEA_SQLITE_PATH", filepath.Join(os.TempDir(), "tea_accrual_redislock_"+utils.GenerateUID()+".db"))
	if err := config.LoadConfig("../../configs/config.yaml"); err != nil {
		t.Fatalf("load config: %v", err)
	}
	// 尝试连接 Redis（根据配置地址）
	database.InitRedis()
	r := database.GetRedis()
	if r == nil {
		t.Skip("redis not available, skip lock test")
	}
	// 初始化 DB
	database.InitDatabase()
	db := database.GetDB()
	u := model.User{BaseModel: model.BaseModel{UID: utils.GenerateUID()}, OpenID: "u_" + utils.GenerateUID(), Phone: "p_" + utils.GenerateUID(), Nickname: "rl", Status: 1, Balance: decimal.NewFromFloat(1000)}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	cfg := config.Accrual{UseRedisLock: true, LockTTLSecond: 60}
	// 使用次日日期，避免与潜在线上/其他测试留下的同日锁冲突
	date := time.Now().Add(24 * time.Hour)

	// 1) 先手动占有锁，然后调用 runOnce，应被跳过（0 条记录）
	if ok := acquireLock(r, date, 30); !ok {
		t.Skip("unable to acquire test lock, skip")
	}
	runOnce(date, 0.01, cfg)
	var cnt int64
	if err := db.Model(&model.InterestRecord{}).Where("user_id = ? AND date(date) = date(?)", u.ID, date.Format("2006-01-02")).Count(&cnt).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if cnt != 0 {
		t.Fatalf("expected 0 with pre-held lock, got %d", cnt)
	}
	releaseLock(r, date)

	// 2) 无锁情况下，直接调用服务层应产生 1 条记录
	if n, err := service.NewAccrualService().Run(date, 0.01, &u.ID); err != nil {
		t.Fatalf("service run: %v", err)
	} else if n == 0 {
		t.Fatalf("service run updated 0 users")
	}
	cnt = 0
	if err := db.Model(&model.InterestRecord{}).Where("user_id = ?", u.ID).Count(&cnt).Error; err != nil {
		t.Fatalf("count2: %v", err)
	}
	if cnt < 1 {
		t.Fatalf("expect >=1 record after unlocked run, got %d", cnt)
	}
}
