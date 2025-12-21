package model

// Brand 品牌模型
type Brand struct {
	BaseModel
	Name            string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	LogoURL         string `gorm:"type:varchar(512)" json:"logo_url"`
	OriginRegionID  *uint  `gorm:"index" json:"origin_region_id"`
	Description     string `gorm:"type:text" json:"description"`
}
