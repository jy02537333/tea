package repo

import (
	"context"
	"errors"
	"tea-api/pkg/database"
)

type WalletRepository struct{}

func NewWalletRepository() *WalletRepository { return &WalletRepository{} }

// Balance returns wallet balance in Yuan by converting cents from DB to float64.
func (r *WalletRepository) Balance(ctx context.Context, userID int64) (float64, error) {
	// Preferred: wallets.balance (cents)
	type walletRow struct{ Balance int64 }
	var row walletRow
	err := database.GetDB().WithContext(ctx).Table("wallets").Select("balance").Where("user_id = ?", userID).Take(&row).Error
	if err == nil {
		return float64(row.Balance) / 100.0, nil
	}
	// Fallback: users.balance (decimal in Yuan)
	type userRow struct{ Balance float64 }
	var u userRow
	err2 := database.GetDB().WithContext(ctx).Table("users").Select("balance").Where("id = ?", userID).Take(&u).Error
	if err2 == nil {
		return u.Balance, nil
	}
	return 0, errors.New("wallet balance unavailable")
}
