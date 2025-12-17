package commission

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

func setupCommissionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Commission{}, &model.CommissionTransaction{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	database.DB = db
	return db
}

func TestReleaseFrozenCommissions(t *testing.T) {
	db := setupCommissionTestDB(t)
	past := time.Now().Add(-48 * time.Hour)
	future := time.Now().Add(48 * time.Hour)

	rec1 := model.Commission{
		UserID:           1,
		CommissionType:   "direct",
		Status:           StatusFrozen,
		Rate:             decimal.NewFromFloat(0.30),
		CalculationBasis: decimal.NewFromInt(1000),
		GrossAmount:      decimal.NewFromInt(300),
		NetAmount:        decimal.NewFromInt(270),
		AvailableAt:      &past,
	}
	rec2 := model.Commission{
		UserID:           2,
		CommissionType:   "direct",
		Status:           StatusFrozen,
		Rate:             decimal.NewFromFloat(0.30),
		CalculationBasis: decimal.NewFromInt(500),
		GrossAmount:      decimal.NewFromInt(150),
		NetAmount:        decimal.NewFromInt(135),
		AvailableAt:      &future,
	}

	if err := db.Create(&rec1).Error; err != nil {
		t.Fatalf("create rec1: %v", err)
	}
	if err := db.Create(&rec2).Error; err != nil {
		t.Fatalf("create rec2: %v", err)
	}

	processed, err := ReleaseFrozenCommissions(10)
	if err != nil {
		t.Fatalf("release err: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected 1 commission released, got %d", processed)
	}

	var updated model.Commission
	if err := db.First(&updated, rec1.ID).Error; err != nil {
		t.Fatalf("get updated rec1: %v", err)
	}
	if updated.Status != StatusAvailable {
		t.Fatalf("rec1 should be available, got %s", updated.Status)
	}

	var txCount int64
	if err := db.Model(&model.CommissionTransaction{}).
		Where("commission_id = ? AND type = ?", rec1.ID, "release").
		Count(&txCount).Error; err != nil {
		t.Fatalf("count release tx: %v", err)
	}
	if txCount != 1 {
		t.Fatalf("expected 1 release tx, got %d", txCount)
	}

	var untouched model.Commission
	if err := db.First(&untouched, rec2.ID).Error; err != nil {
		t.Fatalf("get rec2: %v", err)
	}
	if untouched.Status != StatusFrozen {
		t.Fatalf("rec2 should remain frozen, got %s", untouched.Status)
	}
}

func TestMarkCommissionWithdrawn(t *testing.T) {
	db := setupCommissionTestDB(t)

	rec := model.Commission{
		UserID:           10,
		CommissionType:   "direct",
		Status:           StatusAvailable,
		Rate:             decimal.NewFromFloat(0.40),
		CalculationBasis: decimal.NewFromInt(2000),
		GrossAmount:      decimal.NewFromInt(800),
		NetAmount:        decimal.NewFromInt(800),
	}
	if err := db.Create(&rec).Error; err != nil {
		t.Fatalf("create rec: %v", err)
	}

	operator := uint(99)
	processed, err := MarkCommissionWithdrawn([]uint{rec.ID}, "WD-UNITTEST", &operator, "batch-1")
	if err != nil {
		t.Fatalf("withdraw err: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected 1 commission withdrawn, got %d", processed)
	}

	var updated model.Commission
	if err := db.First(&updated, rec.ID).Error; err != nil {
		t.Fatalf("query commission: %v", err)
	}
	if updated.Status != StatusPaid {
		t.Fatalf("commission should be paid, got %s", updated.Status)
	}

	var txn model.CommissionTransaction
	if err := db.Where("commission_id = ? AND type = ?", rec.ID, "withdraw").First(&txn).Error; err != nil {
		t.Fatalf("query withdraw tx: %v", err)
	}
	if txn.OperatorID == nil || *txn.OperatorID != operator {
		t.Fatalf("expected operator id %d, got %v", operator, txn.OperatorID)
	}
	if txn.ExternalTxnID != "WD-UNITTEST" {
		t.Fatalf("unexpected external txn id: %s", txn.ExternalTxnID)
	}
	if !txn.BalanceAfter.IsZero() {
		t.Fatalf("balance after withdraw should be zero, got %s", txn.BalanceAfter)
	}
}

func TestMarkCommissionWithdrawn_SkipInvalidStatus(t *testing.T) {
	db := setupCommissionTestDB(t)

	good := model.Commission{
		UserID:           1,
		CommissionType:   "direct",
		Status:           StatusAvailable,
		Rate:             decimal.NewFromFloat(0.30),
		CalculationBasis: decimal.NewFromInt(1000),
		GrossAmount:      decimal.NewFromInt(300),
		NetAmount:        decimal.NewFromInt(300),
	}
	bad := model.Commission{
		UserID:           2,
		CommissionType:   "direct",
		Status:           StatusFrozen,
		Rate:             decimal.NewFromFloat(0.30),
		CalculationBasis: decimal.NewFromInt(500),
		GrossAmount:      decimal.NewFromInt(150),
		NetAmount:        decimal.NewFromInt(150),
	}
	if err := db.Create(&good).Error; err != nil {
		t.Fatalf("create good commission: %v", err)
	}
	if err := db.Create(&bad).Error; err != nil {
		t.Fatalf("create bad commission: %v", err)
	}

	processed, err := MarkCommissionWithdrawn([]uint{good.ID, bad.ID}, "WD-FAIL", nil, "")
	if err != nil {
		t.Fatalf("unexpected error when withdrawing with mixed statuses: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected only 1 commission processed, got %d", processed)
	}

	var refreshedGood model.Commission
	if err := db.First(&refreshedGood, good.ID).Error; err != nil {
		t.Fatalf("query good commission: %v", err)
	}
	if refreshedGood.Status != StatusPaid {
		t.Fatalf("good commission status should be paid, got %s", refreshedGood.Status)
	}

	var refreshedBad model.Commission
	if err := db.First(&refreshedBad, bad.ID).Error; err != nil {
		t.Fatalf("query bad commission: %v", err)
	}
	if refreshedBad.Status != StatusFrozen {
		t.Fatalf("bad commission status should remain frozen, got %s", refreshedBad.Status)
	}

	var txCount int64
	if err := db.Model(&model.CommissionTransaction{}).
		Where("commission_id = ? AND type = ?", good.ID, "withdraw").
		Count(&txCount).Error; err != nil {
		t.Fatalf("count withdraw tx: %v", err)
	}
	if txCount != 1 {
		t.Fatalf("expected 1 withdraw tx for good commission, got %d", txCount)
	}
}

func TestConsumeUserAvailableCommissions(t *testing.T) {
	db := setupCommissionTestDB(t)

	userID := uint(42)

	// 三条已解冻佣金，按 ID 升序模拟时间顺序
	c1 := model.Commission{
		UserID:           userID,
		CommissionType:   "direct",
		Status:           StatusAvailable,
		NetAmount:        decimal.NewFromInt(100),
		GrossAmount:      decimal.NewFromInt(100),
		CalculationBasis: decimal.NewFromInt(1000),
	}
	c2 := model.Commission{
		UserID:           userID,
		CommissionType:   "direct",
		Status:           StatusAvailable,
		NetAmount:        decimal.NewFromInt(200),
		GrossAmount:      decimal.NewFromInt(200),
		CalculationBasis: decimal.NewFromInt(2000),
	}
	c3 := model.Commission{
		UserID:           userID,
		CommissionType:   "direct",
		Status:           StatusAvailable,
		NetAmount:        decimal.NewFromInt(300),
		GrossAmount:      decimal.NewFromInt(300),
		CalculationBasis: decimal.NewFromInt(3000),
	}
	if err := db.Create(&c1).Error; err != nil {
		t.Fatalf("create c1: %v", err)
	}
	if err := db.Create(&c2).Error; err != nil {
		t.Fatalf("create c2: %v", err)
	}
	if err := db.Create(&c3).Error; err != nil {
		t.Fatalf("create c3: %v", err)
	}

	// 目标金额 250，应消费 c1 和 c2，两条记录，总金额 300
	processed, consumed, err := ConsumeUserAvailableCommissions(userID, decimal.NewFromInt(250), "WD-AMT", nil, "")
	if err != nil {
		t.Fatalf("consume commissions err: %v", err)
	}
	if processed != 2 {
		t.Fatalf("expected 2 commissions processed, got %d", processed)
	}
	if !consumed.Equal(decimal.NewFromInt(300)) {
		t.Fatalf("expected consumed amount 300, got %s", consumed)
	}

	var updated1, updated2, updated3 model.Commission
	if err := db.First(&updated1, c1.ID).Error; err != nil {
		t.Fatalf("query c1: %v", err)
	}
	if err := db.First(&updated2, c2.ID).Error; err != nil {
		t.Fatalf("query c2: %v", err)
	}
	if err := db.First(&updated3, c3.ID).Error; err != nil {
		t.Fatalf("query c3: %v", err)
	}
	if updated1.Status != StatusPaid || updated2.Status != StatusPaid {
		t.Fatalf("c1 and c2 should be paid, got %s and %s", updated1.Status, updated2.Status)
	}
	if updated3.Status != StatusAvailable {
		t.Fatalf("c3 should remain available, got %s", updated3.Status)
	}

	// 确认只对 c1、c2 生成 withdraw 流水
	var cnt int64
	if err := db.Model(&model.CommissionTransaction{}).
		Where("commission_id IN ? AND type = ?", []uint{c1.ID, c2.ID}, "withdraw").
		Count(&cnt).Error; err != nil {
		t.Fatalf("count withdraw tx: %v", err)
	}
	if cnt != 2 {
		t.Fatalf("expected 2 withdraw transactions, got %d", cnt)
	}
	if err := db.Model(&model.CommissionTransaction{}).
		Where("commission_id = ? AND type = ?", c3.ID, "withdraw").
		Count(&cnt).Error; err != nil {
		t.Fatalf("count c3 withdraw tx: %v", err)
	}
	if cnt != 0 {
		t.Fatalf("expected 0 withdraw tx for c3, got %d", cnt)
	}
}

func TestReverseOrderCommissions(t *testing.T) {
	db := setupCommissionTestDB(t)

	orderID := uint(1001)

	frozen := model.Commission{
		UserID:           1,
		OrderID:          &orderID,
		CommissionType:   "direct",
		Status:           StatusFrozen,
		NetAmount:        decimal.NewFromInt(100),
		GrossAmount:      decimal.NewFromInt(100),
		CalculationBasis: decimal.NewFromInt(1000),
	}
	available := model.Commission{
		UserID:           1,
		OrderID:          &orderID,
		CommissionType:   "direct",
		Status:           StatusAvailable,
		NetAmount:        decimal.NewFromInt(200),
		GrossAmount:      decimal.NewFromInt(200),
		CalculationBasis: decimal.NewFromInt(2000),
	}
	paid := model.Commission{
		UserID:           1,
		OrderID:          &orderID,
		CommissionType:   "direct",
		Status:           StatusPaid,
		NetAmount:        decimal.NewFromInt(300),
		GrossAmount:      decimal.NewFromInt(300),
		CalculationBasis: decimal.NewFromInt(3000),
	}
	if err := db.Create(&frozen).Error; err != nil {
		t.Fatalf("create frozen: %v", err)
	}
	if err := db.Create(&available).Error; err != nil {
		t.Fatalf("create available: %v", err)
	}
	if err := db.Create(&paid).Error; err != nil {
		t.Fatalf("create paid: %v", err)
	}

	op := uint(99)
	processed, err := ReverseOrderCommissions(orderID, &op, "order refund rollback")
	if err != nil {
		t.Fatalf("reverse commissions err: %v", err)
	}
	if processed != 2 {
		t.Fatalf("expected 2 commissions reversed, got %d", processed)
	}

	var rf, ra, rp model.Commission
	if err := db.First(&rf, frozen.ID).Error; err != nil {
		t.Fatalf("query frozen: %v", err)
	}
	if err := db.First(&ra, available.ID).Error; err != nil {
		t.Fatalf("query available: %v", err)
	}
	if err := db.First(&rp, paid.ID).Error; err != nil {
		t.Fatalf("query paid: %v", err)
	}
	if rf.Status != StatusReversed || ra.Status != StatusReversed {
		t.Fatalf("frozen/available should be reversed, got %s and %s", rf.Status, ra.Status)
	}
	if rp.Status != StatusPaid {
		t.Fatalf("paid commission should remain paid, got %s", rp.Status)
	}

	// 确认只针对两条记录生成 adjust 流水
	var cnt int64
	if err := db.Model(&model.CommissionTransaction{}).
		Where("commission_id IN ? AND type = ?", []uint{frozen.ID, available.ID}, "adjust").
		Count(&cnt).Error; err != nil {
		t.Fatalf("count adjust tx: %v", err)
	}
	if cnt != 2 {
		t.Fatalf("expected 2 adjust transactions, got %d", cnt)
	}
	if err := db.Model(&model.CommissionTransaction{}).
		Where("commission_id = ? AND type = ?", paid.ID, "adjust").
		Count(&cnt).Error; err != nil {
		t.Fatalf("count paid adjust tx: %v", err)
	}
	if cnt != 0 {
		t.Fatalf("expected 0 adjust tx for paid commission, got %d", cnt)
	}
}
