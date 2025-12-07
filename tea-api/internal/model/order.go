package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Order 订单模型
type Order struct {
	BaseModel
	OrderNo        string          `gorm:"type:varchar(32);uniqueIndex;not null" json:"order_no"`
	UserID         uint            `gorm:"index;not null" json:"user_id"`
	StoreID        uint            `gorm:"index;default:0" json:"store_id"`
	TotalAmount    decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	PayAmount      decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"pay_amount"`
	DiscountAmount decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	DeliveryFee    decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"delivery_fee"`
	Status         int             `gorm:"type:tinyint;default:1" json:"status"`        // 1:待付款 2:已付款 3:配送中 4:已完成 5:已取消
	PayStatus      int             `gorm:"type:tinyint;default:1" json:"pay_status"`    // 1:未付款 2:已付款 3:退款中 4:已退款
	OrderType      int             `gorm:"type:tinyint;default:1" json:"order_type"`    // 1:商城 2:堂食 3:外卖
	DeliveryType   int             `gorm:"type:tinyint;default:1" json:"delivery_type"` // 1:自取 2:配送
	DeliveryTime   *time.Time      `json:"delivery_time"`
	AddressInfo    string          `gorm:"type:json" json:"address_info"`
	Remark         string          `gorm:"type:text" json:"remark"`
	PaidAt         *time.Time      `json:"paid_at"`
	DeliveredAt    *time.Time      `json:"delivered_at"`
	CompletedAt    *time.Time      `json:"completed_at"`
	CancelledAt    *time.Time      `json:"cancelled_at"`
	CancelReason   string          `gorm:"type:varchar(200)" json:"cancel_reason"`

	User User `gorm:"foreignKey:UserID"`
}

// BeforeSave ensures the JSON column always holds valid content even if client submits blanks.
func (o *Order) BeforeSave(tx *gorm.DB) error {
	o.AddressInfo = NormalizeJSONOrNull(o.AddressInfo)
	return nil
}

// OrderItem 订单项模型
type OrderItem struct {
	BaseModel
	OrderID     uint            `gorm:"index;not null" json:"order_id"`
	ProductID   uint            `gorm:"index;not null" json:"product_id"`
	SkuID       *uint           `gorm:"index" json:"sku_id"`
	ProductName string          `gorm:"type:varchar(100);not null" json:"product_name"`
	SkuName     string          `gorm:"type:varchar(100)" json:"sku_name"`
	Price       decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	Quantity    int             `gorm:"not null" json:"quantity"`
	Amount      decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"amount"`
	Image       string          `gorm:"type:varchar(500)" json:"image"`

	Order   Order      `gorm:"foreignKey:OrderID"`
	Product Product    `gorm:"foreignKey:ProductID"`
	Sku     ProductSku `gorm:"foreignKey:SkuID"`
}

// Cart 购物车模型
type Cart struct {
	BaseModel
	UserID uint `gorm:"index;not null" json:"user_id"`

	User User `gorm:"foreignKey:UserID"`
}

// CartItem 购物车项目模型
type CartItem struct {
	BaseModel
	CartID    uint  `gorm:"index;not null" json:"cart_id"`
	ProductID uint  `gorm:"index;not null" json:"product_id"`
	SkuID     *uint `gorm:"index" json:"sku_id"`
	Quantity  int   `gorm:"not null" json:"quantity"`

	Cart    Cart       `gorm:"foreignKey:CartID"`
	Product Product    `gorm:"foreignKey:ProductID"`
	Sku     ProductSku `gorm:"foreignKey:SkuID"`
}
