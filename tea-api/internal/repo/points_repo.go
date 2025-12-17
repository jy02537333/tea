package repo

import (
	"context"
	"errors"
	"tea-api/pkg/database"
)

type PointsRepository struct{}

func NewPointsRepository() *PointsRepository { return &PointsRepository{} }

// Total returns accumulated usable points by summing points_transactions.change.
func (r *PointsRepository) Total(ctx context.Context, userID int64) (int64, error) {
	var total int64
	err := database.GetDB().WithContext(ctx).Table("points_transactions").
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(`change`),0)").
		Scan(&total).Error
	if err == nil {
		return total, nil
	}
	// Fallback: users.points
	type userRow struct{ Points int64 }
	var u userRow
	err2 := database.GetDB().WithContext(ctx).Table("users").Select("points").Where("id = ?", userID).Take(&u).Error
	if err2 == nil {
		return u.Points, nil
	}
	return 0, errors.New("points unavailable")
}
