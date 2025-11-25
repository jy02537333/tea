package model

import "github.com/shopspring/decimal"

// StoreProduct 门店商品库存与价格覆盖
type StoreProduct struct {
	BaseModel
	StoreID       uint            `gorm:"index:idx_store_product,unique;not null" json:"store_id"`
	ProductID     uint            `gorm:"index:idx_store_product,unique;not null" json:"product_id"`
	Stock         int             `gorm:"not null;default:0" json:"stock"`
	PriceOverride decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"price_override"`
}
