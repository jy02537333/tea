package model

import "github.com/shopspring/decimal"

// StoreProduct 门店商品库存、价格与业务类型覆盖
type StoreProduct struct {
	BaseModel
	StoreID       uint            `gorm:"index:idx_store_product,unique;not null" json:"store_id"`
	ProductID     uint            `gorm:"index:idx_store_product,unique;not null" json:"product_id"`
	Stock         int             `gorm:"not null;default:0" json:"stock"`
	PriceOverride decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"price_override"`
	BizType       int             `gorm:"type:tinyint;not null;default:1" json:"biz_type"` // 1:服务 2:外卖 3:其他
	Product       *Product        `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}
