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

// StoreBankAccount 门店收款账户（用于门店提现打款）
type StoreBankAccount struct {
	BaseModel
	StoreID     uint   `gorm:"index;not null" json:"store_id"`
	AccountType string `gorm:"type:varchar(32);default:'bank'" json:"account_type"`
	AccountName string `gorm:"type:varchar(128)" json:"account_name"`
	AccountNo   string `gorm:"type:varchar(128)" json:"account_no"`
	BankName    string `gorm:"type:varchar(128)" json:"bank_name"`
	IsDefault   bool   `gorm:"default:false" json:"is_default"`
}
