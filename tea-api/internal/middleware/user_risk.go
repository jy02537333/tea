package middleware

import (
	"errors"

	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

type userRiskState struct {
	Status        int
	IsBlacklisted bool
	IsWhitelisted bool
}

func getUserRiskState(userID uint) (*userRiskState, error) {
	db := database.GetDB()
	if db == nil {
		return nil, nil
	}

	var st userRiskState
	err := db.Model(&model.User{}).
		Select("status", "is_blacklisted", "is_whitelisted").
		Where("id = ?", userID).
		Take(&st).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// IsUserBlocked returns whether the user should be blocked from authenticated operations.
// Rule (PRD 3.2.4 黑白名单管理):
// - whitelist: always allow
// - status=2: block
// - blacklist: block
// When DB is not available (e.g., some test setups), it returns allow.
func IsUserBlocked(userID uint) (bool, string) {
	st, err := getUserRiskState(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, "用户不存在"
		}
		return true, "用户状态读取失败"
	}
	if st == nil {
		return false, ""
	}
	if st.IsWhitelisted {
		return false, ""
	}
	if st.Status == 2 {
		return true, "账号已停用"
	}
	if st.IsBlacklisted {
		return true, "账号已被加入黑名单"
	}
	return false, ""
}
