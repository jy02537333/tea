package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var Config = new(config)

type config struct {
	Server        Server        `mapstructure:"server" json:"server" yaml:"server"`
	Database      Database      `mapstructure:"database" json:"database" yaml:"database"`
	Redis         Redis         `mapstructure:"redis" json:"redis" yaml:"redis"`
	RabbitMQ      RabbitMQ      `mapstructure:"rabbitmq" json:"rabbitmq" yaml:"rabbitmq"`
	JWT           JWT           `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Log           Log           `mapstructure:"log" json:"log" yaml:"log"`
	WeChat        WeChat        `mapstructure:"wechat" json:"wechat" yaml:"wechat"`
	Alipay        Alipay        `mapstructure:"alipay" json:"alipay" yaml:"alipay"`
	Delivery      Delivery      `mapstructure:"delivery" json:"delivery" yaml:"delivery"`
	Upload        Upload        `mapstructure:"upload" json:"upload" yaml:"upload"`
	System        System        `mapstructure:"system" json:"system" yaml:"system"`
	Finance       Finance       `mapstructure:"finance" json:"finance" yaml:"finance"`
	Observability Observability `mapstructure:"observability" json:"observability" yaml:"observability"`
	AI            AI            `mapstructure:"ai" json:"ai" yaml:"ai"`
}

type Server struct {
	Port         string        `mapstructure:"port" json:"port" yaml:"port"`
	Mode         string        `mapstructure:"mode" json:"mode" yaml:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" json:"write_timeout" yaml:"write_timeout"`
}

type Database struct {
	Host            string `mapstructure:"host" json:"host" yaml:"host"`
	Port            int    `mapstructure:"port" json:"port" yaml:"port"`
	Username        string `mapstructure:"username" json:"username" yaml:"username"`
	Password        string `mapstructure:"password" json:"password" yaml:"password"`
	DBName          string `mapstructure:"dbname" json:"dbname" yaml:"dbname"`
	Charset         string `mapstructure:"charset" json:"charset" yaml:"charset"`
	ParseTime       bool   `mapstructure:"parse_time" json:"parse_time" yaml:"parse_time"`
	Loc             string `mapstructure:"loc" json:"loc" yaml:"loc"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns" json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns" json:"max_open_conns" yaml:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime" json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
}

func (d Database) Dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		d.Username, d.Password, d.Host, d.Port, d.DBName, d.Charset, d.ParseTime, d.Loc)
}

type Redis struct {
	Host         string `mapstructure:"host" json:"host" yaml:"host"`
	Port         int    `mapstructure:"port" json:"port" yaml:"port"`
	Password     string `mapstructure:"password" json:"password" yaml:"password"`
	DB           int    `mapstructure:"db" json:"db" yaml:"db"`
	PoolSize     int    `mapstructure:"pool_size" json:"pool_size" yaml:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns" json:"min_idle_conns" yaml:"min_idle_conns"`
}

func (r Redis) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type RabbitMQ struct {
	Host              string           `mapstructure:"host" json:"host" yaml:"host"`
	Port              int              `mapstructure:"port" json:"port" yaml:"port"`
	Username          string           `mapstructure:"username" json:"username" yaml:"username"`
	Password          string           `mapstructure:"password" json:"password" yaml:"password"`
	Vhost             string           `mapstructure:"vhost" json:"vhost" yaml:"vhost"`
	ConnectionTimeout int              `mapstructure:"connection_timeout" json:"connection_timeout" yaml:"connection_timeout"`
	Heartbeat         int              `mapstructure:"heartbeat" json:"heartbeat" yaml:"heartbeat"`
	Exchange          RabbitMQExchange `mapstructure:"exchange" json:"exchange" yaml:"exchange"`
	Queues            RabbitMQQueues   `mapstructure:"queues" json:"queues" yaml:"queues"`
}

type RabbitMQExchange struct {
	TeaShop string `mapstructure:"tea_shop" json:"tea_shop" yaml:"tea_shop"`
}

type RabbitMQQueues struct {
	Order         string `mapstructure:"order" json:"order" yaml:"order"`
	Payment       string `mapstructure:"payment" json:"payment" yaml:"payment"`
	Notification  string `mapstructure:"notification" json:"notification" yaml:"notification"`
	ExternalOrder string `mapstructure:"external_order" json:"external_order" yaml:"external_order"`
}

func (r RabbitMQ) DSN() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d%s", r.Username, r.Password, r.Host, r.Port, r.Vhost)
}

type JWT struct {
	Secret      string `mapstructure:"secret" json:"secret" yaml:"secret"`
	ExpiresTime int64  `mapstructure:"expires_time" json:"expires_time" yaml:"expires_time"`
	BufferTime  int64  `mapstructure:"buffer_time" json:"buffer_time" yaml:"buffer_time"`
	Issuer      string `mapstructure:"issuer" json:"issuer" yaml:"issuer"`
}

type Log struct {
	Level      string `mapstructure:"level" json:"level" yaml:"level"`
	Filename   string `mapstructure:"filename" json:"filename" yaml:"filename"`
	MaxSize    int    `mapstructure:"max_size" json:"max_size" yaml:"max_size"`
	MaxAge     int    `mapstructure:"max_age" json:"max_age" yaml:"max_age"`
	MaxBackups int    `mapstructure:"max_backups" json:"max_backups" yaml:"max_backups"`
	Compress   bool   `mapstructure:"compress" json:"compress" yaml:"compress"`
}

type WeChat struct {
	AppID     string `mapstructure:"app_id" json:"app_id" yaml:"app_id"`
	AppSecret string `mapstructure:"app_secret" json:"app_secret" yaml:"app_secret"`
	MchID     string `mapstructure:"mch_id" json:"mch_id" yaml:"mch_id"`
	APIKey    string `mapstructure:"api_key" json:"api_key" yaml:"api_key"`
	CertPath  string `mapstructure:"cert_path" json:"cert_path" yaml:"cert_path"`
	KeyPath   string `mapstructure:"key_path" json:"key_path" yaml:"key_path"`
	NotifyURL string `mapstructure:"notify_url" json:"notify_url" yaml:"notify_url"`
}

type Alipay struct {
	AppID      string `mapstructure:"app_id" json:"app_id" yaml:"app_id"`
	PrivateKey string `mapstructure:"private_key" json:"private_key" yaml:"private_key"`
	PublicKey  string `mapstructure:"public_key" json:"public_key" yaml:"public_key"`
	NotifyURL  string `mapstructure:"notify_url" json:"notify_url" yaml:"notify_url"`
}

type Delivery struct {
	Meituan DeliveryPlatform `mapstructure:"meituan" json:"meituan" yaml:"meituan"`
	Eleme   DeliveryPlatform `mapstructure:"eleme" json:"eleme" yaml:"eleme"`
	Baidu   DeliveryPlatform `mapstructure:"baidu" json:"baidu" yaml:"baidu"`
}

type DeliveryPlatform struct {
	AppKey    string `mapstructure:"app_key" json:"app_key" yaml:"app_key"`
	AppSecret string `mapstructure:"app_secret" json:"app_secret" yaml:"app_secret"`
	BaseURL   string `mapstructure:"base_url" json:"base_url" yaml:"base_url"`
}

type Upload struct {
	Local Local `mapstructure:"local" json:"local" yaml:"local"`
	OSS   OSS   `mapstructure:"oss" json:"oss" yaml:"oss"`
}

type Local struct {
	Path string `mapstructure:"path" json:"path" yaml:"path"`
}

type OSS struct {
	Endpoint        string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id" json:"access_key_id" yaml:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret" json:"access_key_secret" yaml:"access_key_secret"`
	BucketName      string `mapstructure:"bucket_name" json:"bucket_name" yaml:"bucket_name"`
}

type System struct {
	Env           string `mapstructure:"env" json:"env" yaml:"env"`
	Addr          string `mapstructure:"addr" json:"addr" yaml:"addr"`
	DbType        string `mapstructure:"db_type" json:"db_type" yaml:"db_type"`
	UseMultipoint bool   `mapstructure:"use_multipoint" json:"use_multipoint" yaml:"use_multipoint"`
	UseRedis      bool   `mapstructure:"use_redis" json:"use_redis" yaml:"use_redis"`
	LimitCount    int    `mapstructure:"iplimit_count" json:"iplimit_count" yaml:"iplimit_count"`
	LimitTime     int    `mapstructure:"iplimit_time" json:"iplimit_time" yaml:"iplimit_time"`
}

// Finance 财务相关配置
type Finance struct {
	Accrual           Accrual           `mapstructure:"accrual" json:"accrual" yaml:"accrual"`
	CommissionRelease CommissionRelease `mapstructure:"commission_release" json:"commission_release" yaml:"commission_release"`
	Withdrawal        Withdrawal        `mapstructure:"withdrawal" json:"withdrawal" yaml:"withdrawal"`
}

type Accrual struct {
	Enabled       bool     `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Time          string   `mapstructure:"time" json:"time" yaml:"time"` // HH:MM (24h)
	Rate          float64  `mapstructure:"rate" json:"rate" yaml:"rate"` // daily rate (e.g., 0.001)
	UseRedisLock  bool     `mapstructure:"use_redis_lock" json:"use_redis_lock" yaml:"use_redis_lock"`
	LockTTLSecond int      `mapstructure:"lock_ttl_second" json:"lock_ttl_second" yaml:"lock_ttl_second"`
	AllowedRoles  []string `mapstructure:"allowed_roles" json:"allowed_roles" yaml:"allowed_roles"` // who can trigger accrual
	Timezone      string   `mapstructure:"timezone" json:"timezone" yaml:"timezone"`
	SkipWeekends  bool     `mapstructure:"skip_weekends" json:"skip_weekends" yaml:"skip_weekends"`
	Holidays      []string `mapstructure:"holidays" json:"holidays" yaml:"holidays"`
}

// CommissionRelease 佣金解冻调度配置
type CommissionRelease struct {
	Enabled       bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Time          string `mapstructure:"time" json:"time" yaml:"time"` // HH:MM (24h)
	UseRedisLock  bool   `mapstructure:"use_redis_lock" json:"use_redis_lock" yaml:"use_redis_lock"`
	LockTTLSecond int    `mapstructure:"lock_ttl_second" json:"lock_ttl_second" yaml:"lock_ttl_second"`
	Timezone      string `mapstructure:"timezone" json:"timezone" yaml:"timezone"`
	BatchSize     int    `mapstructure:"batch_size" json:"batch_size" yaml:"batch_size"`
}

// Withdrawal 提现费用与限额配置
type Withdrawal struct {
	MinAmountCents int64 `mapstructure:"min_amount_cents" json:"min_amount_cents" yaml:"min_amount_cents"`
	FeeFixedCents  int64 `mapstructure:"fee_fixed_cents" json:"fee_fixed_cents" yaml:"fee_fixed_cents"`
	FeeRateBp      int64 `mapstructure:"fee_rate_bp" json:"fee_rate_bp" yaml:"fee_rate_bp"` // 手续费比例，基点（万分制），如 30 表示 0.30%
	FeeMinCents    int64 `mapstructure:"fee_min_cents" json:"fee_min_cents" yaml:"fee_min_cents"`
	FeeCapCents    int64 `mapstructure:"fee_cap_cents" json:"fee_cap_cents" yaml:"fee_cap_cents"`
}

// Observability 可观测性配置
type Observability struct {
	OperationLog OperationLog `mapstructure:"operationlog" json:"operationlog" yaml:"operationlog"`
}

type OperationLog struct {
	Enabled         bool     `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	IncludePrefixes []string `mapstructure:"include_prefixes" json:"include_prefixes" yaml:"include_prefixes"`
	ExcludePrefixes []string `mapstructure:"exclude_prefixes" json:"exclude_prefixes" yaml:"exclude_prefixes"`
}

// AI runtime configuration
type AI struct {
	Enabled        bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Provider       string `mapstructure:"provider" json:"provider" yaml:"provider"`
	DefaultModel   string `mapstructure:"default_model" json:"default_model" yaml:"default_model"`
	APIKeyEnv      string `mapstructure:"api_key_env" json:"api_key_env" yaml:"api_key_env"`
	APIURL         string `mapstructure:"api_url" json:"api_url" yaml:"api_url"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds" json:"timeout_seconds" yaml:"timeout_seconds"`
	MaxConcurrency int    `mapstructure:"max_concurrency" json:"max_concurrency" yaml:"max_concurrency"`
}

// LoadConfig 加载配置文件
func LoadConfig(path string) error {
	viper.SetConfigFile(path)
	// 默认开启操作日志
	viper.SetDefault("observability.operationlog.enabled", true)
	// 默认 AI 不启用
	viper.SetDefault("ai.enabled", false)
	// support environment overrides. Prefix env vars with TEA_, e.g. TEA_DATABASE_HOST
	viper.SetEnvPrefix("TEA")
	viper.AutomaticEnv()

	// Bind commonly overridden environment variables so CI/local scripts can set them easily
	_ = viper.BindEnv("database.host", "TEA_DATABASE_HOST")
	_ = viper.BindEnv("database.port", "TEA_DATABASE_PORT")
	_ = viper.BindEnv("database.username", "TEA_DATABASE_USERNAME")
	_ = viper.BindEnv("database.password", "TEA_DATABASE_PASSWORD")
	_ = viper.BindEnv("database.dbname", "TEA_DATABASE_DBNAME")

	_ = viper.BindEnv("redis.host", "TEA_REDIS_HOST")
	_ = viper.BindEnv("redis.port", "TEA_REDIS_PORT")
	_ = viper.BindEnv("redis.password", "TEA_REDIS_PASSWORD")

	_ = viper.BindEnv("rabbitmq.host", "TEA_RABBITMQ_HOST")
	_ = viper.BindEnv("rabbitmq.port", "TEA_RABBITMQ_PORT")
	_ = viper.BindEnv("rabbitmq.username", "TEA_RABBITMQ_USERNAME")
	_ = viper.BindEnv("rabbitmq.password", "TEA_RABBITMQ_PASSWORD")
	_ = viper.BindEnv("rabbitmq.vhost", "TEA_RABBITMQ_VHOST")

	// Set safe defaults for test/dev environment so tests use the shared endpoints
	viper.SetDefault("database.host", "127.0.0.1")
	viper.SetDefault("database.port", 3308)
	viper.SetDefault("database.username", "root")
	viper.SetDefault("database.password", "gs963852")
	viper.SetDefault("database.dbname", "tea_shop")
	viper.SetDefault("database.charset", "utf8mb4")
	viper.SetDefault("database.parse_time", true)
	viper.SetDefault("database.loc", "Local")

	viper.SetDefault("redis.host", "127.0.0.1")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "123456")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("rabbitmq.host", "127.0.0.1")
	viper.SetDefault("rabbitmq.port", 5672)
	viper.SetDefault("rabbitmq.username", "guest")
	viper.SetDefault("rabbitmq.password", "guest")
	viper.SetDefault("rabbitmq.vhost", "/")

	// Withdrawal defaults
	viper.SetDefault("finance.withdrawal.min_amount_cents", 1000) // 最低提现 10 元
	viper.SetDefault("finance.withdrawal.fee_fixed_cents", 0)     // 固定手续费（默认 0）
	viper.SetDefault("finance.withdrawal.fee_rate_bp", 30)        // 比例手续费：0.30%
	viper.SetDefault("finance.withdrawal.fee_min_cents", 100)     // 最低手续费 1 元
	viper.SetDefault("finance.withdrawal.fee_cap_cents", 0)       // 封顶手续费（0 表示不封顶）

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&Config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// If a full DSN is provided via TEA_DSN, try to parse and override database fields.
	// Expected DSN format: user:pass@tcp(host:port)/dbname?params
	if dsn := os.Getenv("TEA_DSN"); dsn != "" {
		// strip query params
		parts := strings.SplitN(dsn, "?", 2)
		core := parts[0]

		// regex capture: user, pass, host, port, dbname
		re := regexp.MustCompile(`^([^:]+):([^@]+)@tcp\(([^:]+):(\d+)\)/([^/]+)$`)
		if m := re.FindStringSubmatch(core); len(m) == 6 {
			Config.Database.Username = m[1]
			Config.Database.Password = m[2]
			Config.Database.Host = m[3]
			if p, err := strconv.Atoi(m[4]); err == nil {
				Config.Database.Port = p
			}
			Config.Database.DBName = m[5]
		}
		// Note: if parsing fails we silently keep existing config (from file/env bindings)
	}

	return nil
}
