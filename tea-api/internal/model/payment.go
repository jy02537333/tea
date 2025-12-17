package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Payment 支付记录模型
type Payment struct {
	BaseModel
	OrderID       uint            `gorm:"index;not null" json:"order_id"`
	PaymentNo     string          `gorm:"type:varchar(64);uniqueIndex;not null" json:"payment_no"`
	PaymentMethod int             `gorm:"type:tinyint;not null" json:"payment_method"` // 1:微信 2:支付宝
	Amount        decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status        int             `gorm:"type:tinyint;default:1" json:"status"` // 1:待支付 2:支付成功 3:支付失败
	ThirdPayNo    string          `gorm:"type:varchar(64)" json:"third_pay_no"`
	ThirdResponse string          `gorm:"type:text" json:"third_response"`
	PaidAt        *time.Time      `json:"paid_at"`
	NotifyAt      *time.Time      `json:"notify_at"`

	Order Order `gorm:"foreignKey:OrderID"`
}

// Refund 退款记录模型
type Refund struct {
	BaseModel
	OrderID       uint            `gorm:"index;not null" json:"order_id"`
	PaymentID     uint            `gorm:"index;not null" json:"payment_id"`
	RefundNo      string          `gorm:"type:varchar(64);uniqueIndex;not null" json:"refund_no"`
	RefundAmount  decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"refund_amount"`
	RefundReason  string          `gorm:"type:varchar(200)" json:"refund_reason"`
	Status        int             `gorm:"type:tinyint;default:1" json:"status"` // 1:申请中 2:退款成功 3:退款失败
	ThirdRefundNo string          `gorm:"type:varchar(64)" json:"third_refund_no"`
	ThirdResponse string          `gorm:"type:text" json:"third_response"`
	RefundedAt    *time.Time      `json:"refunded_at"`

	Order   Order   `gorm:"foreignKey:OrderID"`
	Payment Payment `gorm:"foreignKey:PaymentID"`
}

// 提现状态常量
const (
	WithdrawStatusPending    = 1 // 申请中
	WithdrawStatusProcessing = 2 // 处理中
	WithdrawStatusCompleted  = 3 // 已完成
	WithdrawStatusRejected   = 4 // 已拒绝
)

// WithdrawRecord 提现记录模型
type WithdrawRecord struct {
	BaseModel
	UserID       uint            `gorm:"index;not null" json:"user_id"`
	StoreID      uint            `gorm:"index;default:0" json:"store_id"`
	WithdrawNo   string          `gorm:"type:varchar(64);uniqueIndex;not null" json:"withdraw_no"`
	Amount       decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"amount"`
	Fee          decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"fee"`
	ActualAmount decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"actual_amount"`
	WithdrawType int             `gorm:"type:tinyint;not null" json:"withdraw_type"` // 1:微信转账
	Status       int             `gorm:"type:tinyint;default:1" json:"status"`       // 参见 WithdrawStatus* 常量
	Remark       string          `gorm:"type:varchar(200)" json:"remark"`
	ProcessedAt  *time.Time      `json:"processed_at"`
	ProcessedBy  uint            `json:"processed_by"`

	User User `gorm:"foreignKey:UserID"`
}

// WechatTransferRecord 微信转账记录模型
type WechatTransferRecord struct {
	BaseModel
	WithdrawID     uint            `gorm:"index;not null" json:"withdraw_id"`
	PartnerTradeNo string          `gorm:"type:varchar(64);uniqueIndex;not null" json:"partner_trade_no"`
	OpenID         string          `gorm:"type:varchar(50);not null" json:"open_id"`
	Amount         decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"amount"`
	Description    string          `gorm:"type:varchar(100)" json:"description"`
	Status         int             `gorm:"type:tinyint;default:1" json:"status"` // 1:处理中 2:转账成功 3:转账失败
	PaymentNo      string          `gorm:"type:varchar(64)" json:"payment_no"`
	PaymentTime    *time.Time      `json:"payment_time"`
	ErrorCode      string          `gorm:"type:varchar(20)" json:"error_code"`
	ErrorMsg       string          `gorm:"type:varchar(200)" json:"error_msg"`

	Withdraw WithdrawRecord `gorm:"foreignKey:WithdrawID"`
}
