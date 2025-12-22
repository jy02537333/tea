package service

import (
	"context"
	"testing"
)

// fake repos for unit tests
type fakeWallet struct{ v float64 }
func (f fakeWallet) Balance(ctx context.Context, userID int64) (float64, error) { return f.v, nil }
type fakePoints struct{ v int64 }
func (f fakePoints) Total(ctx context.Context, userID int64) (int64, error) { return f.v, nil }
type fakeCoupons struct{ v int }
func (f fakeCoupons) UsableCount(ctx context.Context, userID int64) (int, error) { return f.v, nil }
type fakeMembership struct{ v string }
func (f fakeMembership) Level(ctx context.Context, userID int64) (string, error) { return f.v, nil }

func TestGetSummary_EmptyDeps(t *testing.T) {
    s := NewUserSummaryServiceWithDeps(SummaryDeps{})
    sum, err := s.GetSummary(context.Background(), 1)
    if err != nil { t.Fatalf("err: %v", err) }
    if sum.WalletBalance != 0 || sum.Points != 0 || sum.Coupons != 0 || sum.Membership != "visitor" {
        t.Fatalf("unexpected defaults: %+v", sum)
    }
}

func TestGetSummary_PartialDeps(t *testing.T) {
    s := NewUserSummaryServiceWithDeps(SummaryDeps{Wallet: fakeWallet{v: 12.34}})
    sum, err := s.GetSummary(context.Background(), 1)
    if err != nil { t.Fatalf("err: %v", err) }
    if sum.WalletBalance != 12.34 || sum.Points != 0 || sum.Coupons != 0 || sum.Membership != "visitor" {
        t.Fatalf("unexpected partial: %+v", sum)
    }
}

func TestGetSummary_FullDeps(t *testing.T) {
    s := NewUserSummaryServiceWithDeps(SummaryDeps{
        Wallet:     fakeWallet{v: 56.78},
        Points:     fakePoints{v: 150},
        Coupons:    fakeCoupons{v: 2},
        Membership: fakeMembership{v: "gold"},
    })
    sum, err := s.GetSummary(context.Background(), 1)
    if err != nil { t.Fatalf("err: %v", err) }
    if sum.WalletBalance != 56.78 || sum.Points != 150 || sum.Coupons != 2 || sum.Membership != "gold" {
        t.Fatalf("unexpected full: %+v", sum)
    }
}
