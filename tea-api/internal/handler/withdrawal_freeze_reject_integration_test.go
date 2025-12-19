//go:build integration

package handler

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

// local ensureWallet using gorm.DB
func ensureWallet2(db *gorm.DB, userID uint, balance, frozen int64) error {
	var cnt int64
	if err := db.Table("wallets").Where("user_id = ?", userID).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return db.Exec("INSERT INTO wallets (user_id, balance, frozen) VALUES (?,?,?)", userID, balance, frozen).Error
	}
	return db.Exec("UPDATE wallets SET balance=?, frozen=? WHERE user_id=?", balance, frozen, userID).Error
}

func TestWithdraw_FreezeThenRejectFlow(t *testing.T) {
	db, err := database.InitWithoutMigrate()
	if err != nil || db == nil {
		t.Skipf("skip: db unavailable: %v", err)
		return
	}

	userID := uint(900003)
	// seed wallet: balance=200000, frozen=0
	if err := ensureWallet2(db, userID, 200000, 0); err != nil {
		t.Fatalf("seed wallet failed: %v", err)
	}

	// step1: user apply -> freeze 100000
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", userID)
	c.Request = httptest.NewRequest("POST", "/api/v1/wallet/withdrawals", strings.NewReader(`{"bank_account_id":0,"amount_cents":100000,"currency":"CNY","note":"test"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	CreateMyWithdrawal(c)
	if w.Code != 200 {
		t.Fatalf("withdraw apply failed: status=%d body=%s", w.Code, w.Body.String())
	}

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

	// step2: create an admin-side withdraw record corresponding to the frozen amount
	rec := &model.WithdrawRecord{
		UserID:       userID,
		WithdrawNo:   "WD_TEST_FR_REJ",
		Amount:       decimal.NewFromInt(1000), // 100000 cents
		Fee:          decimal.NewFromInt(3),    // 300 cents
		ActualAmount: decimal.NewFromInt(997),  // 99700 cents
		Status:       model.WithdrawStatusPending,
	}
	if err := db.Create(rec).Error; err != nil {
		t.Fatalf("create admin withdraw record failed: %v", err)
	}

	// step3: admin reject -> unfreeze
	admin := NewWithdrawAdminHandler()
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Set("user_id", uint(1))
	c2.Params = gin.Params{gin.Param{Key: "id", Value: fmt.Sprintf("%d", rec.ID)}}
	c2.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/withdraws/%d/reject", rec.ID), strings.NewReader(`{"remark":"integration reject"}`))
	c2.Request.Header.Set("Content-Type", "application/json")
	admin.Reject(c2)
	if w2.Code != 200 {
		t.Fatalf("admin reject failed: status=%d body=%s", w2.Code, w2.Body.String())
	}

	// verify wallet after reject: balance=200000, frozen=0
	if err := db.Table("wallets").Select("balance,frozen").Where("user_id = ?", userID).Take(&wal).Error; err != nil {
		t.Fatalf("query wallet after reject failed: %v", err)
	}
	if wal.Balance != 200000 || wal.Frozen != 0 {
		t.Fatalf("unexpected wallet after reject: balance=%d frozen=%d", wal.Balance, wal.Frozen)
	}

	// verify last transaction JSON remark
	var txRow struct {
		Type   string
		Remark string
	}
	if err := db.Table("wallet_transactions").Select("type,remark").Where("user_id = ?", userID).Order("id desc").Take(&txRow).Error; err != nil {
		t.Fatalf("query last tx failed: %v", err)
	}
	if txRow.Type != "withdraw_reject_unfreeze" {
		t.Fatalf("unexpected tx type: %s", txRow.Type)
	}
	var rmk map[string]any
	if err := json.Unmarshal([]byte(txRow.Remark), &rmk); err != nil {
		t.Fatalf("remark not JSON: %v, raw=%s", err, txRow.Remark)
	}
	if rmk["phase"] != "rejected_unfreeze" {
		t.Fatalf("remark.phase unexpected: %v", rmk["phase"])
	}
}
