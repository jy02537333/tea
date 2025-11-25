package model

// Store 门店模型
type Store struct {
	BaseModel
	Name          string  `gorm:"type:varchar(100);not null" json:"name"`
	Address       string  `gorm:"type:varchar(300)" json:"address"`
	Phone         string  `gorm:"type:varchar(20)" json:"phone"`
	Latitude      float64 `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude     float64 `gorm:"type:decimal(10,7)" json:"longitude"`
	BusinessHours string  `gorm:"type:varchar(200)" json:"business_hours"`
	Images        string  `gorm:"type:text" json:"images"`
	Status        int     `gorm:"type:tinyint;default:1" json:"status"` // 1启用 2停业
}
