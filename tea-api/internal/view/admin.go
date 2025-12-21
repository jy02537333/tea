package view

import (
	"github.com/shopspring/decimal"
)

// CategoryLite 为 Admin 列表使用的轻量级分类对象
type CategoryLite struct {
	Name string `json:"name,omitempty"`
}

// AdminProductListItem 统一的 Admin 产品列表视图模型
// 保持字段精简，并为门店维度兼容可选字段。
// 注意：价格字段使用 string（shopspring/decimal 的 JSON 表示），以便前端直接展示。
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
package view




























}	StorePriceOverride *string        `json:"store_price_override,omitempty"`	StoreStock        *int            `json:"store_stock,omitempty"`	BrandName         *string         `json:"brand_name,omitempty"`	BrandID           *uint           `json:"brand_id,omitempty"`	UpdatedAt         string          `json:"updated_at,omitempty"`	Stock             int             `json:"stock"`	Status            int             `json:"status"`	Price             decimal.Decimal `json:"price"`	Category          *CategoryLite   `json:"category,omitempty"`	CategoryID        uint            `json:"category_id"`	Name              string          `json:"name"`	ID                uint            `json:"id"`type AdminProductListItem struct {// 注意：价格字段使用 string（shopspring/decimal 的 JSON 表示），以便前端直接展示。// 保持字段精简，并为门店维度兼容可选字段。// AdminProductListItem 统一的 Admin 产品列表视图模型}	Name string `json:"name,omitempty"`type CategoryLite struct {// CategoryLite 为 Admin 列表使用的轻量级分类对象)	"github.com/shopspring/decimal"import (package view