package repo

import (
	"context"
	"tea-api/pkg/database"
)

type MembershipRepository struct{}

func NewMembershipRepository() *MembershipRepository { return &MembershipRepository{} }

// Level returns current membership package name for active record; empty if none.
func (r *MembershipRepository) Level(ctx context.Context, userID int64) (string, error) {
	type row struct{ Name string }
	var res row
	err := database.GetDB().WithContext(ctx).
		Table("user_memberships AS um").
		Select("mp.name AS name").
		Joins("JOIN membership_packages mp ON mp.id = um.package_id").
		Where("um.user_id = ? AND um.status = 'active'", userID).
		Order("um.started_at DESC").
		Limit(1).
		Scan(&res).Error
	if err != nil {
		return "", err
	}
	return res.Name, nil
}
