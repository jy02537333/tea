package repo

import (
	"context"
	"tea-api/pkg/database"
	"time"
)

type CouponsRepository struct{}

func NewCouponsRepository() *CouponsRepository { return &CouponsRepository{} }

// UsableCount returns count of user's coupons with status=unused and not expired.
func (r *CouponsRepository) UsableCount(ctx context.Context, userID int64) (int, error) {
	var cnt int64
	now := time.Now()
	err := database.GetDB().WithContext(ctx).Table("coupons").
		Where("user_id = ? AND status = 'unused' AND (expires_at IS NULL OR expires_at > ?)", userID, now).
		Count(&cnt).Error
	return int(cnt), err
}
