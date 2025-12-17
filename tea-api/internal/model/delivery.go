package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// DeliveryOrder 配送订单模型
type DeliveryOrder struct {
	BaseModel
	OrderID         uint            `gorm:"index;not null" json:"order_id"`
	Platform        int             `gorm:"type:tinyint;not null" json:"platform"` // 1:美团 2:饿了么 3:百度
	PlatformOrderID string          `gorm:"type:varchar(100);not null" json:"platform_order_id"`
	DeliveryFee     decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"delivery_fee"`
	Status          int             `gorm:"type:tinyint;default:1" json:"status"` // 1:待接单 2:已接单 3:配送中 4:已送达 5:已取消
	DeliveryTime    *time.Time      `json:"delivery_time"`
	DeliveredAt     *time.Time      `json:"delivered_at"`

	Order Order `gorm:"foreignKey:OrderID"`
}

// DeliveryPlatformOrder 外卖平台订单模型
type DeliveryPlatformOrder struct {
	BaseModel
	Platform        int        `gorm:"type:tinyint;not null" json:"platform"` // 1:美团 2:饿了么 3:百度
	PlatformOrderID string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"platform_order_id"`
	OrderData       string     `gorm:"type:text" json:"order_data"`               // 平台订单原始数据JSON
	SyncStatus      int        `gorm:"type:tinyint;default:1" json:"sync_status"` // 1:待同步 2:已同步 3:同步失败
	LocalOrderID    uint       `gorm:"index" json:"local_order_id"`
	SyncAt          *time.Time `json:"sync_at"`
	ErrorMsg        string     `gorm:"type:varchar(500)" json:"error_msg"`

	LocalOrder Order `gorm:"foreignKey:LocalOrderID"`
}

// Coupon 优惠券模型
type Coupon struct {
	BaseModel
	// StoreID 为空表示平台券，非空表示门店券
	StoreID     *uint           `gorm:"index" json:"store_id"`
	Name        string          `gorm:"type:varchar(100);not null" json:"name"`
	Type        int             `gorm:"type:tinyint;not null" json:"type"` // 1:满减券 2:折扣券 3:免单券
	Amount      decimal.Decimal `gorm:"type:decimal(10,2)" json:"amount"`
	Discount    decimal.Decimal `gorm:"type:decimal(3,2)" json:"discount"`              // 折扣率，如0.8表示8折
	MinAmount   decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"min_amount"` // 最低消费金额
	TotalCount  int             `gorm:"not null" json:"total_count"`
	UsedCount   int             `gorm:"default:0" json:"used_count"`
	Status      int             `gorm:"type:tinyint;default:1" json:"status"` // 1:启用 2:禁用
	StartTime   time.Time       `gorm:"not null" json:"start_time"`
	EndTime     time.Time       `gorm:"not null" json:"end_time"`
	Description string          `gorm:"type:text" json:"description"`
}

// UserCoupon 用户优惠券模型
type UserCoupon struct {
	BaseModel
	UserID   uint       `gorm:"index;not null" json:"user_id"`
	CouponID uint       `gorm:"index;not null" json:"coupon_id"`
	OrderID  *uint      `gorm:"index" json:"order_id"`
	Status   int        `gorm:"type:tinyint;default:1" json:"status"` // 1:未使用 2:已使用 3:已过期
	UsedAt   *time.Time `json:"used_at"`

	User   User   `gorm:"foreignKey:UserID"`
	Coupon Coupon `gorm:"foreignKey:CouponID"`
	Order  Order  `gorm:"foreignKey:OrderID"`
}

// Activity 营销活动模型
type Activity struct {
	BaseModel
	// StoreID 为空表示平台活动，非空表示门店活动
	StoreID     *uint     `gorm:"index" json:"store_id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Type        int       `gorm:"type:tinyint;not null" json:"type"` // 1:限时折扣 2:满减活动 3:买赠活动
	StartTime   time.Time `gorm:"not null" json:"start_time"`
	EndTime     time.Time `gorm:"not null" json:"end_time"`
	Rules       string    `gorm:"type:json" json:"rules"`               // 活动规则JSON
	Status      int       `gorm:"type:tinyint;default:1" json:"status"` // 1:启用 2:禁用
	Priority    int       `gorm:"default:0" json:"priority"`
	Description string    `gorm:"type:text" json:"description"`
}

// ActivityProduct 活动商品模型
type ActivityProduct struct {
	BaseModel
	ActivityID uint `gorm:"index;not null" json:"activity_id"`
	ProductID  uint `gorm:"index;not null" json:"product_id"`

	Activity Activity `gorm:"foreignKey:ActivityID"`
	Product  Product  `gorm:"foreignKey:ProductID"`
}

// ActivityRegistration 活动报名记录
// 简化版：金额与退款仅做记录，不直接驱动支付/退款流水（后续可与订单/退款模块打通）
type ActivityRegistration struct {
	BaseModel
	StoreID      uint            `gorm:"index;not null" json:"store_id"`
	ActivityID   uint            `gorm:"index;not null" json:"activity_id"`
	UserID       uint            `gorm:"index;not null" json:"user_id"`
	UserName     string          `gorm:"type:varchar(100)" json:"user_name"`
	UserPhone    string          `gorm:"type:varchar(20)" json:"user_phone"`
	OrderID      *uint           `gorm:"index" json:"order_id"`
	Status       int             `gorm:"type:tinyint;default:1" json:"status"` // 1:已报名（待支付/免费） 2:已支付报名 3:已退款
	OrderStatus  int             `gorm:"-" json:"order_status,omitempty"`
	PayStatus    int             `gorm:"-" json:"order_pay_status,omitempty"`
	Fee          decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"fee"`           // 报名费用
	RefundAmount decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"refund_amount"` // 实际退款金额
	RefundReason string          `gorm:"type:varchar(255)" json:"refund_reason"`
	RefundedAt   *time.Time      `json:"refunded_at"`
}
