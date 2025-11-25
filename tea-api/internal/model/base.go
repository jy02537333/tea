package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"tea-api/pkg/utils"
)

// BaseModel 基础模型，包含审计字段
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UID       string         `gorm:"type:varchar(32);uniqueIndex;not null" json:"uid"`
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy uint           `gorm:"index" json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy uint           `gorm:"index" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	IsDeleted bool           `gorm:"default:false" json:"is_deleted"`
}

// BeforeCreate 自动填充 UID
func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if b.UID == "" {
		b.UID = utils.GenerateUID()
	}
	return nil
}

// SoftDeleteModel 软删除模型
type SoftDeleteModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UID       string         `gorm:"type:varchar(32);uniqueIndex;not null" json:"uid"`
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy uint           `gorm:"index" json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy uint           `gorm:"index" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// User 用户模型
type User struct {
	BaseModel
	Username     string          `gorm:"type:varchar(100);uniqueIndex" json:"username"`
	PasswordHash string          `gorm:"type:varchar(255)" json:"-"`
	OpenID       string          `gorm:"type:varchar(50);uniqueIndex" json:"open_id"`
	UnionID      string          `gorm:"type:varchar(50);index" json:"union_id"`
	Phone        string          `gorm:"type:varchar(20);uniqueIndex" json:"phone"`
	Nickname     string          `gorm:"type:varchar(50)" json:"nickname"`
	Avatar       string          `gorm:"type:varchar(500)" json:"avatar"`
	Gender       int             `gorm:"type:tinyint;default:0" json:"gender"` // 0:未知 1:男 2:女
	Birthday     *time.Time      `json:"birthday"`
	Province     string          `gorm:"type:varchar(50)" json:"province"`
	City         string          `gorm:"type:varchar(50)" json:"city"`
	Country      string          `gorm:"type:varchar(50)" json:"country"`
	Status       int             `gorm:"type:tinyint;default:1" json:"status"` // 1:正常 2:禁用
	LastLoginAt  *time.Time      `json:"last_login_at"`
	Balance      decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"balance"`
	InterestRate decimal.Decimal `gorm:"type:decimal(8,6);default:0" json:"interest_rate"` // 用户定制日利率，>0 时覆盖默认
	Points       int             `gorm:"default:0" json:"points"`
	Role         string          `gorm:"type:varchar(30);default:'user';index" json:"role"` // 简化角色标识，详尽权限通过关联表
}

// Role 角色模型
type Role struct {
	BaseModel
	Name        string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	DisplayName string `gorm:"type:varchar(100)" json:"display_name"`
	Description string `gorm:"type:text" json:"description"`
	Status      int    `gorm:"type:tinyint;default:1" json:"status"` // 1:启用 2:禁用
}

// Permission 权限模型
type Permission struct {
	BaseModel
	Name        string `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	DisplayName string `gorm:"type:varchar(100)" json:"display_name"`
	Description string `gorm:"type:text" json:"description"`
	Module      string `gorm:"type:varchar(50)" json:"module"`
	Action      string `gorm:"type:varchar(50)" json:"action"`
	Resource    string `gorm:"type:varchar(100)" json:"resource"`
}

// UserRole 用户角色关联模型
type UserRole struct {
	BaseModel
	UserID uint `gorm:"index;not null" json:"user_id"`
	RoleID uint `gorm:"index;not null" json:"role_id"`

	User User `gorm:"foreignKey:UserID"`
	Role Role `gorm:"foreignKey:RoleID"`
}

// RolePermission 角色权限关联模型
type RolePermission struct {
	BaseModel
	RoleID       uint `gorm:"index;not null" json:"role_id"`
	PermissionID uint `gorm:"index;not null" json:"permission_id"`

	Role       Role       `gorm:"foreignKey:RoleID"`
	Permission Permission `gorm:"foreignKey:PermissionID"`
}
