//go:build integration

package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tea-api/pkg/database"
)

// helper: ensure wallet row exists
func ensureWallet(db *gorm.DB, userID uint, balance, frozen int64) error {
	var cnt int64
	if err := db.Table("wallets").Where("user_id = ?", userID).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return db.Exec("INSERT INTO wallets (user_id, balance, frozen) VALUES (?,?,?)", userID, balance, frozen).Error
	}
	return db.Exec("UPDATE wallets SET balance=?, frozen=? WHERE user_id=?", balance, frozen, userID).Error
}

func TestWithdraw_FreezeAndPaidFlow(t *testing.T) {
	// initialize DB without migration; skip test if unavailable
	db, err := database.InitWithoutMigrate()
	if err != nil || db == nil {
		t.Skipf("skip: db unavailable: %v", err)
		return
	}

	userID := uint(900001)
	// seed wallet: 200000 cents balance, 0 frozen
	if err := ensureWallet(db, userID, 200000, 0); err != nil {
		t.Fatalf("seed wallet failed: %v", err)
	}

	// call CreateMyWithdrawal via gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", userID)
	c.Request = httptest.NewRequest("POST", "/api/v1/wallet/withdrawals", strings.NewReader(`{"bank_account_id":0,"amount_cents":100000,"currency":"CNY","note":"test"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	CreateMyWithdrawal(c)
	if w.Code != 200 {
		t.Fatalf("withdraw apply failed: status=%d body=%s", w.Code, w.Body.String())
	}

	// verify wallet frozen changed and balance reduced
	var wal struct {
		Balance int64
		Frozen  int64
	}
	if err := db.Table("wallets").Select("balance,frozen").Where("user_id = ?", userID).Take(&wal).Error; err != nil {
		t.Fatalf("query wallet after freeze failed: %v", err)
	}
	if wal.Balance != 100000 || wal.Frozen != 100000 {
		t.Fatalf("unexpected wallet after freeze: balance=%d frozen=%d", wal.Balance, wal.Frozen)
	}

	// simulate admin paid: finalize (unfreeze)
	// build a minimal WithdrawRecord compatible object
	// we avoid importing internal model here to keep test lightweight; directly exec SQL for wallet_transactions
	// Expected: frozen decreases by 100000, balance unchanged
	if err := db.Exec("UPDATE wallets SET frozen = frozen - ? WHERE user_id = ?", 100000, userID).Error; err != nil {
		t.Fatalf("simulate paid failed: %v", err)
	}
	if err := db.Exec("INSERT INTO wallet_transactions (user_id, type, amount, balance_after, remark, created_at) VALUES (?,?,?,?,?,CURRENT_TIMESTAMP)", userID, "withdraw_paid", 0, wal.Balance, "integration finalize").Error; err != nil {
		t.Fatalf("record finalize tx failed: %v", err)
	}

	if err := db.Table("wallets").Select("balance,frozen").Where("user_id = ?", userID).Take(&wal).Error; err != nil {
		t.Fatalf("query wallet after paid failed: %v", err)
	}
	if wal.Balance != 100000 || wal.Frozen != 0 {
		t.Fatalf("unexpected wallet after paid: balance=%d frozen=%d", wal.Balance, wal.Frozen)
	}
}
