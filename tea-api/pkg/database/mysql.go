package database

import (
	"fmt"
	"log"
	"os"
	"tea-test/pkg/env"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"tea-api/internal/config"
	"tea-api/internal/model"
)

var DB *gorm.DB

// InitMySQL 初始化MySQL数据库
func InitMySQL() {
	cfg := config.Config.Database

	fmt.Printf("正在连接MySQL数据库: %s@%s:%d/%s\n", cfg.Username, cfg.Host, cfg.Port, cfg.DBName)

	// 配置日志
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, // 改为Info级别以便调试
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound错误
			Colorful:                  true,        // 启用彩色打印
		},
	)

	dsn := cfg.Dsn()
	fmt.Printf("数据库连接字符串: %s\n", dsn)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("数据库连接失败: %v\n", err)
		panic(fmt.Errorf("failed to connect database: %w", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Errorf("failed to get underlying sql.DB: %w", err))
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	DB = db

	fmt.Println("数据库连接池配置完成")

	// 启动时自动执行 GORM 迁移，保持 schema 与模型同步。
	// 若迁移失败，仅打印错误并继续启动，避免影响线上可用性。
	fmt.Println("开始执行数据库自动迁移...")
	if err := autoMigrate(); err != nil {
		fmt.Printf("数据库自动迁移失败，但继续启动服务器: %v\n", err)
	} else {
		fmt.Println("数据库自动迁移完成!")
	}
}

// autoMigrate 自动迁移数据表
func autoMigrate() error {
	return DB.AutoMigrate(
		// 基础模型
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},

		// 商品管理
		&model.Category{},
		&model.Product{},
		&model.ProductSku{},
		&model.ProductImage{},

		// 订单管理
		&model.Store{},
		&model.StoreBankAccount{},
		&model.StoreProduct{},
		&model.Order{},
		&model.OrderItem{},
		&model.Cart{},
		&model.CartItem{},

		// 支付管理
		&model.Payment{},
		&model.Refund{},

		// 外卖平台
		&model.DeliveryOrder{},
		&model.DeliveryPlatformOrder{},

		// 营销活动
		&model.Coupon{},
		&model.UserCoupon{},
		&model.Activity{},
		&model.ActivityProduct{},
		&model.ActivityRegistration{},

		// 配置管理
		&model.SystemConfig{},
		&model.Banner{},

		// 统计分析
		&model.OrderStatistics{},
		&model.ProductStatistics{},
		&model.UserStatistics{},

		// 日志管理
		&model.AccessLog{},
		&model.OperationLog{},

		// 资金/利息
		&model.InterestRecord{},
		// 佣金与提现相关
		&model.Commission{},
		&model.CommissionTransaction{},
		&model.Wallet{},
		&model.WalletTransaction{},
		&model.MembershipPackage{},
		&model.PartnerLevel{},
		&model.UserBankAccount{},
		&model.ReferralClosure{},

		// 提现管理
		&model.WithdrawRecord{},
		&model.WithdrawalRequest{},
		&model.WechatTransferRecord{},
	)
}

// InitWithoutMigrate 初始化数据库连接但不执行迁移
func InitWithoutMigrate() (*gorm.DB, error) {
	cfg := config.Config.Database

	fmt.Printf("正在连接MySQL数据库: %s@%s:%d/%s (跳过迁移)\n", cfg.Username, cfg.Host, cfg.Port, cfg.DBName)

	// 配置日志
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // 慢 SQL 阈值
			LogLevel:                  logger.Silent, // 设置为Silent减少日志输出
			IgnoreRecordNotFoundError: true,          // 忽略ErrRecordNotFound错误
			Colorful:                  true,          // 启用彩色打印
		},
	)

	dsn := cfg.Dsn()

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	DB = db

	fmt.Println("✅ 数据库连接成功（跳过迁移模式）")

	return DB, nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return DB
}

// InitDatabase 根据环境变量选择数据库，支持 SQLite 回退
// 当 TEA_USE_SQLITE=1 时，优先使用本地 SQLite 文件。
func InitDatabase() {
	// Allow tests to skip DB initialization to avoid requiring external DB or cgo.
	if env.Get("TEA_SKIP_DB_INIT", "") == "1" {
		fmt.Println("Skipping DB init because TEA_SKIP_DB_INIT=1")
		return
	}
	// 根据环境变量选择是否执行自动迁移
	val := env.Get("TEA_AUTO_MIGRATE", "1")
	fmt.Printf("[InitDatabase] TEA_AUTO_MIGRATE=%s\n", val)
	if val == "0" {
		fmt.Println("[InitDatabase] Using InitWithoutMigrate (skip auto migration)")
		if _, err := InitWithoutMigrate(); err != nil {
			log.Printf("InitWithoutMigrate failed: %v\n", err)
			// 回退到带迁移的初始化以保证可用性
			fmt.Println("[InitDatabase] Fallback to InitMySQL with auto migration")
			InitMySQL()
		}
		return
	}
	fmt.Println("[InitDatabase] Using InitMySQL with auto migration (default)")
	// 默认走 MySQL 初始化（执行自动迁移）
	InitMySQL()
}
