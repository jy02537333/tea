package view

import "github.com/shopspring/decimal"

// CategoryLite 为 Admin 列表使用的轻量级分类对象
type CategoryLite struct {
	Name string `json:"name,omitempty"`
}

// AdminProductListItem 统一的 Admin 产品列表视图模型
// 保持字段精简，并为门店维度兼容可选字段。
// 注意：价格字段使用 decimal，方便前端直接展示字符串值。
type AdminProductListItem struct {
	ID                 uint            `json:"id"`
	Name               string          `json:"name"`
	CategoryID         uint            `json:"category_id"`
	Category           *CategoryLite   `json:"category,omitempty"`
	Price              decimal.Decimal `json:"price"`
	Status             int             `json:"status"`
	Stock              int             `json:"stock"`
	UpdatedAt          string          `json:"updated_at,omitempty"`
	BrandID            *uint           `json:"brand_id,omitempty"`
	BrandName          *string         `json:"brand_name,omitempty"`
	StoreStock         *int            `json:"store_stock,omitempty"`
	StorePriceOverride *string         `json:"store_price_override,omitempty"`
}
