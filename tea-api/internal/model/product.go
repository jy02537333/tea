package model

import (
	"github.com/shopspring/decimal"
)

// Category 商品分类模型
type Category struct {
	BaseModel
	Name        string `gorm:"type:varchar(50);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Image       string `gorm:"type:varchar(500)" json:"image"`
	Sort        int    `gorm:"default:0" json:"sort"`
	Status      int    `gorm:"type:tinyint;default:1" json:"status"` // 1:启用 2:禁用
	ParentID    uint   `gorm:"default:0" json:"parent_id"`
}

// Product 商品模型
type Product struct {
	BaseModel
	CategoryID    uint            `gorm:"index;not null" json:"category_id"`
	BrandID       *uint           `gorm:"index" json:"brand_id"`
	Name          string          `gorm:"type:varchar(100);not null" json:"name"`
	Description   string          `gorm:"type:text" json:"description"`
	Images        string          `gorm:"type:text" json:"images"` // JSON数组
	Price         decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	OriginalPrice decimal.Decimal `gorm:"type:decimal(10,2)" json:"original_price"`
	Stock         int             `gorm:"default:0" json:"stock"`
	Sales         int             `gorm:"default:0" json:"sales"`
	Status        int             `gorm:"type:tinyint;default:1" json:"status"` // 1:上架 2:下架
	Sort          int             `gorm:"default:0" json:"sort"`
	IsHot         bool            `gorm:"default:false" json:"is_hot"`
	IsNew         bool            `gorm:"default:false" json:"is_new"`
	IsRecommend   bool            `gorm:"default:false" json:"is_recommend"`

	Category Category     `gorm:"foreignKey:CategoryID"`
	Brand    *Brand       `gorm:"foreignKey:BrandID" json:"brand"`
	Skus     []ProductSku `gorm:"foreignKey:ProductID" json:"skus"`
}

// ProductSku 商品SKU模型
type ProductSku struct {
	BaseModel
	ProductID uint            `gorm:"index;not null" json:"product_id"`
	SkuName   string          `gorm:"type:varchar(100)" json:"sku_name"`
	SkuCode   string          `gorm:"type:varchar(50);uniqueIndex" json:"sku_code"`
	Price     decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock     int             `gorm:"default:0" json:"stock"`
	Sales     int             `gorm:"default:0" json:"sales"`
	Attrs     string          `gorm:"type:json" json:"attrs"` // 规格属性JSON
	Image     string          `gorm:"type:varchar(500)" json:"image"`
	Status    int             `gorm:"type:tinyint;default:1" json:"status"`

	Product Product `gorm:"foreignKey:ProductID"`
}

// ProductImage 商品图片模型
type ProductImage struct {
	BaseModel
	ProductID uint   `gorm:"index;not null" json:"product_id"`
	ImageURL  string `gorm:"type:varchar(500);not null" json:"image_url"`
	Sort      int    `gorm:"default:0" json:"sort"`
	IsMain    bool   `gorm:"default:false" json:"is_main"`

	Product Product `gorm:"foreignKey:ProductID"`
}
