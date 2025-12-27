package model

import "time"

// WithdrawalRequest 提现申请（对应数据库表 withdrawal_requests）
type WithdrawalRequest struct {
    ID            uint       `gorm:"primaryKey" json:"id"`
    UserID        uint       `gorm:"index;not null" json:"user_id"`
    Amount        int64      `gorm:"not null" json:"amount"` // 单位：分
    Fee           int64      `gorm:"default:0" json:"fee"`
    Status        string     `gorm:"type:varchar(32);default:'pending'" json:"status"`
    BankAccountID *uint      `gorm:"column:bank_account_id" json:"bank_account_id,omitempty"`
    InvoiceRequired bool     `gorm:"column:invoice_required;default:0" json:"invoice_required"`
    InvoiceNo     string     `gorm:"column:invoice_no;type:varchar(128)" json:"invoice_no,omitempty"`
    Remark        string     `gorm:"type:varchar(255)" json:"remark,omitempty"`
    RequestedAt   time.Time  `gorm:"column:requested_at;autoCreateTime" json:"requested_at"`
    ProcessedAt   *time.Time `gorm:"column:processed_at" json:"processed_at,omitempty"`
}
