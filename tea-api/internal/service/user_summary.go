package service

import "context"

type UserSummary struct {
	UserID        int64   `json:"user_id"`
	Nickname      string  `json:"nickname"`
	WalletBalance float64 `json:"wallet_balance"`
	Points        int64   `json:"points"`
	Coupons       int     `json:"coupons"`
	Membership    string  `json:"membership"`
}

// UserSummaryService aggregates user-related info from repos.
type UserSummaryService interface {
	GetSummary(ctx context.Context, userID int64) (UserSummary, error)
}

// Repository interfaces for aggregation (to be implemented elsewhere).
type WalletRepo interface {
	Balance(ctx context.Context, userID int64) (float64, error)
}
type PointsRepo interface {
	Total(ctx context.Context, userID int64) (int64, error)
}
type CouponsRepo interface {
	UsableCount(ctx context.Context, userID int64) (int, error)
}
type MembershipRepo interface {
	Level(ctx context.Context, userID int64) (string, error)
}

type SummaryDeps struct {
	Wallet     WalletRepo
	Points     PointsRepo
	Coupons    CouponsRepo
	Membership MembershipRepo
}

type userSummaryService struct {
	deps SummaryDeps
}

// Backward-compatible constructor (no deps => zero/default values)
func NewUserSummaryService() UserSummaryService { return &userSummaryService{} }

// Preferred constructor with explicit dependencies.
func NewUserSummaryServiceWithDeps(deps SummaryDeps) UserSummaryService {
	return &userSummaryService{deps: deps}
}

func (s *userSummaryService) GetSummary(ctx context.Context, userID int64) (UserSummary, error) {
	var (
		balance float64
		points  int64
		coupons int
		level   = "visitor"
	)

	if s.deps.Wallet != nil {
		if v, err := s.deps.Wallet.Balance(ctx, userID); err == nil {
			balance = v
		}
	}
	if s.deps.Points != nil {
		if v, err := s.deps.Points.Total(ctx, userID); err == nil {
			points = v
		}
	}
	if s.deps.Coupons != nil {
		if v, err := s.deps.Coupons.UsableCount(ctx, userID); err == nil {
			coupons = v
		}
	}
	if s.deps.Membership != nil {
		if v, err := s.deps.Membership.Level(ctx, userID); err == nil && v != "" {
			level = v
		}
	}

	return UserSummary{
		UserID:        userID,
		Nickname:      "访客",
		WalletBalance: balance,
		Points:        points,
		Coupons:       coupons,
		Membership:    level,
	}, nil
}
