package model

import "time"

// SystemConfig 系统配置模型
type SystemConfig struct {
	BaseModel
	ConfigKey   string `gorm:"type:varchar(100);uniqueIndex;not null" json:"config_key"`
	ConfigValue string `gorm:"type:text" json:"config_value"`
	ConfigType  string `gorm:"type:varchar(20);default:'string'" json:"config_type"` // string, int, float, bool, json
	Description string `gorm:"type:varchar(200)" json:"description"`
	Status      int    `gorm:"type:tinyint;default:1" json:"status"` // 1:启用 2:禁用
}

// Banner 轮播图模型
type Banner struct {
	BaseModel
	Title    string `gorm:"type:varchar(100)" json:"title"`
	ImageURL string `gorm:"type:varchar(500);not null" json:"image_url"`
	LinkType int    `gorm:"type:tinyint;default:1" json:"link_type"` // 1:无链接 2:商品详情 3:分类页 4:外部链接
	LinkURL  string `gorm:"type:varchar(500)" json:"link_url"`
	Sort     int    `gorm:"default:0" json:"sort"`
	Status   int    `gorm:"type:tinyint;default:1" json:"status"` // 1:启用 2:禁用
}

// OrderStatistics 订单统计模型
type OrderStatistics struct {
	BaseModel
	Date        time.Time `gorm:"type:date;uniqueIndex;not null" json:"date"`
	OrderCount  int       `gorm:"default:0" json:"order_count"`
	TotalAmount float64   `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
	AvgAmount   float64   `gorm:"type:decimal(10,2);default:0" json:"avg_amount"`
}

// ProductStatistics 商品统计模型
type ProductStatistics struct {
	BaseModel
	Date       time.Time `gorm:"type:date;not null" json:"date"`
	ProductID  uint      `gorm:"index;not null" json:"product_id"`
	ViewCount  int       `gorm:"default:0" json:"view_count"`
	SalesCount int       `gorm:"default:0" json:"sales_count"`

	Product Product `gorm:"foreignKey:ProductID"`
}

// UserStatistics 用户统计模型
type UserStatistics struct {
	BaseModel
	Date            time.Time `gorm:"type:date;uniqueIndex;not null" json:"date"`
	NewUserCount    int       `gorm:"default:0" json:"new_user_count"`
	ActiveUserCount int       `gorm:"default:0" json:"active_user_count"`
	OrderUserCount  int       `gorm:"default:0" json:"order_user_count"`
}

// AccessLog 访问日志模型
type AccessLog struct {
	BaseModel
	UserID     uint   `gorm:"index" json:"user_id"`
	Method     string `gorm:"type:varchar(10);not null" json:"method"`
	Path       string `gorm:"type:varchar(500);not null" json:"path"`
	Query      string `gorm:"type:text" json:"query"`
	Body       string `gorm:"type:text" json:"body"`
	UserAgent  string `gorm:"type:varchar(500)" json:"user_agent"`
	IP         string `gorm:"type:varchar(50)" json:"ip"`
	StatusCode int    `gorm:"not null" json:"status_code"`
	Latency    int64  `gorm:"not null" json:"latency"` // 响应时间，纳秒

	User User `gorm:"foreignKey:UserID"`
}

// OperationLog 操作日志模型
type OperationLog struct {
	BaseModel
	UserID      uint   `gorm:"index;not null" json:"user_id"`
	Module      string `gorm:"type:varchar(50);not null" json:"module"`
	Operation   string `gorm:"type:varchar(100);not null" json:"operation"`
	Description string `gorm:"type:varchar(500)" json:"description"`
	RequestData string `gorm:"type:text" json:"request_data"`
	IP          string `gorm:"type:varchar(50)" json:"ip"`
	UserAgent   string `gorm:"type:varchar(500)" json:"user_agent"`

	User User `gorm:"foreignKey:UserID"`
}
