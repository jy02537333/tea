package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// InterestRecord 利息计提记录（定点精度）
type InterestRecord struct {
	BaseModel
	UserID          uint            `gorm:"uniqueIndex:uid_date,priority:1;not null" json:"user_id"`
	Date            time.Time       `gorm:"type:date;uniqueIndex:uid_date,priority:2;not null" json:"date"`
	PrincipalBefore decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"principal_before"`
	Rate            decimal.Decimal `gorm:"type:decimal(8,6);not null" json:"rate"`
	InterestAmount  decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"interest_amount"`
	PrincipalAfter  decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"principal_after"`
	Method          string          `gorm:"type:varchar(20);default:'daily'" json:"method"` // daily, manual
	Note            string          `gorm:"type:varchar(255)" json:"note"`

	User User `gorm:"foreignKey:UserID"`
}
