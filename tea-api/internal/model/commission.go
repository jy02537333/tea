package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Commission 佣金记录
type Commission struct {
	BaseModel
	UserID           uint            `gorm:"index;not null" json:"user_id"`
	OrderID          *uint           `gorm:"index" json:"order_id"`
	OrderItemID      *uint           `gorm:"index" json:"order_item_id"`
	PackageID        *uint           `gorm:"index" json:"package_id"`
	LevelID          *uint           `gorm:"index" json:"level_id"`
	CommissionType   string          `gorm:"type:varchar(32);index;default:'direct'" json:"commission_type"` // direct|indirect|upgrade
	SourceUserID     *uint           `gorm:"index" json:"source_user_id"`
	Rate             decimal.Decimal `gorm:"type:decimal(8,6);default:0" json:"rate"`
	CalculationBasis decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"calculation_basis"`
	GrossAmount      decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"gross_amount"`
	Fee              decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"fee"`
	NetAmount        decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"net_amount"`
	Status           string          `gorm:"type:varchar(32);index;default:'frozen'" json:"status"`
	AvailableAt      *time.Time      `json:"available_at"`
}

// CommissionTransaction 佣金流水
type CommissionTransaction struct {
	BaseModel
	CommissionID  uint            `gorm:"index;not null" json:"commission_id"`
	Type          string          `gorm:"type:varchar(32);not null" json:"type"` // release|withdraw|adjust|upgrade_reward|fee
	Amount        decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"amount"`
	BalanceAfter  decimal.Decimal `gorm:"type:decimal(12,2)" json:"balance_after"`
	OperatorID    *uint           `gorm:"index" json:"operator_id"`
	ExternalTxnID string          `gorm:"type:varchar(128)" json:"external_txn_id"`
	Note          string          `gorm:"type:varchar(255)" json:"note"`
}

// MembershipPackage 会员/合伙人礼包配置
type MembershipPackage struct {
	BaseModel
	Name                 string          `gorm:"type:varchar(128);not null" json:"name"`
	Price                decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"price"`
	TeaCoinAward         decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"tea_coin_award"`
	DiscountRate         decimal.Decimal `gorm:"type:decimal(6,4);default:1.00" json:"discount_rate"`
	PurchaseDiscountRate decimal.Decimal `gorm:"type:decimal(6,4);default:1.00" json:"purchase_discount_rate"`
	DirectCommissionRate decimal.Decimal `gorm:"type:decimal(6,4);default:0" json:"direct_commission_rate"`
	TeamCommissionRate   decimal.Decimal `gorm:"type:decimal(6,4);default:0" json:"team_commission_rate"`
	UpgradeRewardRate    decimal.Decimal `gorm:"type:decimal(6,4);default:0" json:"upgrade_reward_rate"`
	Type                 string          `gorm:"type:varchar(64);default:'membership'" json:"type"`
}

// PartnerLevel 合伙人等级策略
type PartnerLevel struct {
	BaseModel
	Name                 string          `gorm:"type:varchar(128);not null" json:"name"`
	PurchaseDiscountRate decimal.Decimal `gorm:"type:decimal(6,4);default:1.00" json:"purchase_discount_rate"`
	DirectCommissionRate decimal.Decimal `gorm:"type:decimal(6,4);default:0" json:"direct_commission_rate"`
	TeamCommissionRate   decimal.Decimal `gorm:"type:decimal(6,4);default:0" json:"team_commission_rate"`
	UpgradeRewardRate    decimal.Decimal `gorm:"type:decimal(6,4);default:0" json:"upgrade_reward_rate"`
	Note                 string          `gorm:"type:varchar(255)" json:"note"`
}

// UserBankAccount 用户提现账户（简化）
type UserBankAccount struct {
	BaseModel
	UserID      uint   `gorm:"index;not null" json:"user_id"`
	AccountType string `gorm:"type:varchar(32);default:'bank'" json:"account_type"`
	AccountName string `gorm:"type:varchar(128)" json:"account_name"`
	AccountNo   string `gorm:"type:varchar(128)" json:"account_no"`
	BankName    string `gorm:"type:varchar(128)" json:"bank_name"`
	IsDefault   bool   `gorm:"default:false" json:"is_default"`
}

// ReferralClosure 推荐闭包表（仅用于 ORM 映射）
type ReferralClosure struct {
	AncestorUserID   uint `gorm:"primaryKey;autoIncrement:false" json:"ancestor_user_id"`
	DescendantUserID uint `gorm:"primaryKey;autoIncrement:false" json:"descendant_user_id"`
	Depth            int  `json:"depth"`
}

// Referral 推荐关系表
type Referral struct {
	BaseModel
	ReferrerUserID  uint   `gorm:"index;not null" json:"referrer_user_id"`
	ReferredUserID  uint   `gorm:"uniqueIndex;not null" json:"referred_user_id"`
	Source          string `gorm:"type:varchar(64)" json:"source"` // 来源：share_link/qrcode/invite_code
}

// Wallet 钱包表
type Wallet struct {
	BaseModel
	UserID   uint            `gorm:"uniqueIndex;not null" json:"user_id"`
	Balance  decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"balance"`
	TeaCoins decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"tea_coins"`
	Frozen   decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"frozen"`
}

// WalletTransaction 钱包流水表
type WalletTransaction struct {
	BaseModel
	UserID       uint            `gorm:"index;not null" json:"user_id"`
	Type         string          `gorm:"type:varchar(64);not null" json:"type"` // recharge/consume/refund/commission/withdraw/tea_coin_award
	Amount       decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"amount"`
	BalanceAfter decimal.Decimal `gorm:"type:decimal(12,2)" json:"balance_after"`
	OrderID      *uint           `gorm:"index" json:"order_id"`
	Description  string          `gorm:"type:varchar(255)" json:"description"`
}

// OrderLog 订单日志表
type OrderLog struct {
	BaseModel
	OrderID     uint   `gorm:"index;not null" json:"order_id"`
	Status      string `gorm:"type:varchar(64);not null" json:"status"`
	Description string `gorm:"type:varchar(255)" json:"description"`
}

