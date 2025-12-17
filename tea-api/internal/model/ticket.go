package model

import "time"

// Ticket 客服/投诉工单模型
// 对应 PRD 3.2.9 客服与投诉管理（平台端）
// 状态枚举：new, pending, waiting_customer, resolved, rejected, closed
// 优先级：low, normal, high
// 类型：consult, order, refund, recharge, complaint, other
// 来源：miniapp_feedback, miniapp_order, store_staff, phone, manual

type Ticket struct {
	BaseModel

	Type        string `gorm:"type:varchar(30);index;not null" json:"type"`
	Source      string `gorm:"type:varchar(30);index;not null" json:"source"`
	UserID      *uint  `gorm:"index" json:"user_id"`
	OrderID     *uint  `gorm:"index" json:"order_id"`
	StoreID     *uint  `gorm:"index" json:"store_id"`
	Title       string `gorm:"type:varchar(200);not null" json:"title"`
	Content     string `gorm:"type:text" json:"content"`
	Attachments string `gorm:"type:text" json:"attachments"`

	Status       string `gorm:"type:varchar(20);index;not null;default:'new'" json:"status"`
	Priority     string `gorm:"type:varchar(20);not null;default:'normal'" json:"priority"`
	AssigneeID   *uint  `gorm:"index" json:"assignee_id"`
	Remark       string `gorm:"type:text" json:"remark"`
	RejectReason string `gorm:"type:text" json:"reject_reason"`

	ResolvedAt *time.Time `json:"resolved_at"`
	ClosedAt   *time.Time `json:"closed_at"`
}
