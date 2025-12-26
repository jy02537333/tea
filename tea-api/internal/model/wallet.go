package model

import (
	"time"
)

// Wallet 用户钱包表（余额以分计）
type Wallet struct {
    BaseModel
    UserID uint  `gorm:"primaryKey;column:user_id" json:"user_id"`
    Balance int64 `gorm:"column:balance;not null;default:0" json:"balance"`
    Frozen  int64 `gorm:"column:frozen;not null;default:0" json:"frozen"`
}

// WalletTransaction 钱包流水（单位：分）
type WalletTransaction struct {
    ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID uint `gorm:"index;not null" json:"user_id"`
    Type string `gorm:"type:varchar(64);not null" json:"type"`
    Amount int64 `gorm:"not null" json:"amount"`
    BalanceAfter *int64 `gorm:"column:balance_after" json:"balance_after,omitempty"`
    Remark string `gorm:"type:varchar(255)" json:"remark"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
