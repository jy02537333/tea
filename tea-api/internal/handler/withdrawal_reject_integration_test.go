//go:build integration

package handler

import (
	"encoding/json"
	"testing"

	"tea-api/internal/model"
	"tea-api/pkg/database"

	"github.com/shopspring/decimal"
)

// ensureWallet helper copied pattern from withdrawal_integration_test.go
func ensureWalletReject(db *database.DBWrapper, userID uint, balance, frozen int64) error {
	var cnt int64
	if err := db.DB.Table("wallets").Where("user_id = ?", userID).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return db.DB.Exec("INSERT INTO wallets (user_id, balance, frozen) VALUES (?,?,?)", userID, balance, frozen).Error
	}
	return db.DB.Exec("UPDATE wallets SET balance=?, frozen=? WHERE user_id=?", balance, frozen, userID).Error
}

func TestWithdraw_RejectUnfreezeJSONRemark(t *testing.T) {
	// initialize DB without migration; skip test if unavailable
	db, err := database.InitWithoutMigrate()
	if err != nil || db == nil {
		t.Skipf("skip: db unavailable: %v", err)
		return
	}

	userID := uint(900002)
	// seed wallet to state after freeze: balance=100000, frozen=100000
	if err := ensureWallet(db.DB, userID, 100000, 100000); err != nil { // reuse ensureWallet from existing test file
		t.Fatalf("seed wallet failed: %v", err)
	}

	// construct a minimal WithdrawRecord (amount=100000 cents => 1000.00, fee=300 cents => 3.00)
	rec := &model.WithdrawRecord{
		UserID:       userID,
		WithdrawNo:   "WD_TEST_REJECT",
		Amount:       decimal.NewFromInt(1000),
		Fee:          decimal.NewFromInt(3),
		ActualAmount: decimal.NewFromInt(997),
	}

	// call rollback (reject) to unfreeze and return funds
	if err := rollbackWalletForWithdrawReject(db.DB, rec); err != nil {
		t.Fatalf("rollback reject failed: %v", err)
	}

	// verify wallet after rollback: balance=200000, frozen=0
	var wal struct {
		Balance int64
		Frozen  int64
	}
	if err := db.DB.Table("wallets").Select("balance,frozen").Where("user_id = ?", userID).Take(&wal).Error; err != nil {
		t.Fatalf("query wallet after reject failed: %v", err)
	}
	if wal.Balance != 200000 || wal.Frozen != 0 {
		t.Fatalf("unexpected wallet after reject: balance=%d frozen=%d", wal.Balance, wal.Frozen)
	}

	// fetch last transaction and validate JSON remark
	var txRow struct {
		Type   string
		Remark string
	}
	if err := db.DB.Table("wallet_transactions").Select("type,remark").Where("user_id = ?", userID).Order("id desc").Take(&txRow).Error; err != nil {
		t.Fatalf("query last tx failed: %v", err)
	}
	if txRow.Type != "withdraw_reject_unfreeze" {
		t.Fatalf("unexpected tx type: %s", txRow.Type)
	}
	var rmk map[string]any
	if err := json.Unmarshal([]byte(txRow.Remark), &rmk); err != nil {
		t.Fatalf("remark is not valid JSON: %v, raw=%s", err, txRow.Remark)
	}
	// basic field checks
	if rmk["phase"] != "rejected_unfreeze" {
		t.Fatalf("remark.phase unexpected: %v", rmk["phase"])
	}
	// amount_cents should be 100000
	if v, ok := rmk["amount_cents"].(float64); !ok || int64(v) != 100000 {
		t.Fatalf("remark.amount_cents unexpected: %v", rmk["amount_cents"])
	}
}
