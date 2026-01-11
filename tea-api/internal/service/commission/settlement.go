package commission

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

const (
	// StatusFrozen indicates the commission is locked until the hold period ends.
	StatusFrozen = "frozen"
	// StatusAvailable means funds passed the hold period and can be withdrawn.
	StatusAvailable = "available"
	// StatusPaid marks commissions that have been tied to a withdrawal/payout.
	StatusPaid = "paid"
	// StatusReversed is reserved for refund/chargeback scenarios.
	StatusReversed = "reversed"
)

const (
	defaultReleaseBatchSize = 100
	maxReleaseBatchSize     = 500
)

// ReleaseFrozenCommissionsTx finds commissions whose hold period expired and marks them as available.
// It operates within the provided transaction.
func ReleaseFrozenCommissionsTx(tx *gorm.DB, batchSize int) (int, error) {
	if tx == nil {
		return 0, errors.New("tx is nil")
	}
	if batchSize <= 0 {
		batchSize = defaultReleaseBatchSize
	}
	if batchSize > maxReleaseBatchSize {
		batchSize = maxReleaseBatchSize
	}

	var candidates []model.Commission
	now := time.Now()
	if err := tx.Where("status = ? AND available_at IS NOT NULL AND available_at <= ?", StatusFrozen, now).
		Order("id").
		Limit(batchSize).
		Find(&candidates).Error; err != nil {
		return 0, err
	}
	if len(candidates) == 0 {
		return 0, nil
	}

	processed := 0
	for _, cm := range candidates {
		if err := tx.Model(&model.Commission{}).
			Where("id = ? AND status = ?", cm.ID, StatusFrozen).
			Update("status", StatusAvailable).Error; err != nil {
			return processed, err
		}

		relTx := model.CommissionTransaction{
			CommissionID: cm.ID,
			Type:         "release",
			Amount:       cm.NetAmount,
			BalanceAfter: cm.NetAmount,
			Note:         "auto release",
		}
		if err := tx.Create(&relTx).Error; err != nil {
			return processed, err
		}
		processed++
	}
	return processed, nil
}

// ReleaseFrozenCommissions finds commissions whose hold period expired and marks them as available.
// It also appends a commission transaction entry for auditing purposes.
func ReleaseFrozenCommissions(batchSize int) (int, error) {
	db, err := requireDB()
	if err != nil {
		return 0, err
	}
	tx := db.Begin()
	processed, err := ReleaseFrozenCommissionsTx(tx, batchSize)
	if err != nil {
		tx.Rollback()
		return processed, err
	}
	if err := tx.Commit().Error; err != nil {
		return processed, err
	}
	return processed, nil
}

// MarkCommissionWithdrawn marks a set of commissions as paid and appends withdrawal transactions.
func MarkCommissionWithdrawn(commissionIDs []uint, externalTxnID string, operatorID *uint, note string) (int, error) {
	if len(commissionIDs) == 0 {
		return 0, errors.New("commissionIDs cannot be empty")
	}
	db, err := requireDB()
	if err != nil {
		return 0, err
	}

	tx := db.Begin()
	processed := 0
	for _, id := range commissionIDs {
		var cm model.Commission
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cm, id).Error; err != nil {
			tx.Rollback()
			return processed, err
		}
		if cm.Status != StatusAvailable {
			// 跳过非可提现状态的记录，避免部分成功时整体回滚
			continue
		}

		if err := tx.Model(&model.Commission{}).
			Where("id = ?", cm.ID).
			Update("status", StatusPaid).Error; err != nil {
			tx.Rollback()
			return processed, err
		}

		txn := model.CommissionTransaction{
			CommissionID:  cm.ID,
			Type:          "withdraw",
			Amount:        cm.NetAmount,
			BalanceAfter:  decimal.NewFromInt(0),
			ExternalTxnID: externalTxnID,
			Note:          note,
		}
		if operatorID != nil {
			txn.OperatorID = operatorID
		}
		if err := tx.Create(&txn).Error; err != nil {
			tx.Rollback()
			return processed, err
		}
		processed++
	}

	if err := tx.Commit().Error; err != nil {
		return processed, err
	}
	return processed, nil
}

// ConsumeUserAvailableCommissions 按时间顺序消费某用户已解冻佣金，直至覆盖目标金额。
//
// 规则说明：
//   - 仅选取 status=available 的佣金记录，按创建顺序依次消费；
//   - 每条佣金必须整体标记为已打款，不做“部分拆分”；
//   - 当累计消费金额首次达到或超过 targetAmount 即停止（可能会略大于目标值）；
//   - 若所有可用佣金净额之和仍小于 targetAmount，则返回错误并不落库任何提现流水。
//
// 返回值：processed 为成功标记为已打款的佣金条数，consumed 为这些佣金净额之和。
// 实际落库仍复用 MarkCommissionWithdrawn，以保持流水记录一致性。
func ConsumeUserAvailableCommissions(userID uint, targetAmount decimal.Decimal, externalTxnID string, operatorID *uint, note string) (processed int, consumed decimal.Decimal, err error) {
	if userID == 0 {
		return 0, decimal.Zero, errors.New("userID cannot be zero")
	}
	if targetAmount.LessThanOrEqual(decimal.NewFromInt(0)) {
		return 0, decimal.Zero, nil
	}

	db, err := requireDB()
	if err != nil {
		return 0, decimal.Zero, err
	}

	var available []model.Commission
	if err := db.Where("user_id = ? AND status = ?", userID, StatusAvailable).
		Order("id ASC").
		Find(&available).Error; err != nil {
		return 0, decimal.Zero, err
	}
	if len(available) == 0 {
		return 0, decimal.Zero, errors.New("no available commissions for user")
	}

	remaining := targetAmount
	consumed = decimal.Zero
	var ids []uint
	for _, cm := range available {
		if remaining.LessThanOrEqual(decimal.NewFromInt(0)) {
			break
		}
		ids = append(ids, cm.ID)
		remaining = remaining.Sub(cm.NetAmount)
		consumed = consumed.Add(cm.NetAmount)
	}
	if len(ids) == 0 {
		return 0, decimal.Zero, errors.New("insufficient available commissions")
	}
	if consumed.LessThan(targetAmount) {
		return 0, consumed, errors.New("insufficient available commissions to cover target amount")
	}

	processed, err = MarkCommissionWithdrawn(ids, externalTxnID, operatorID, note)
	if err != nil {
		return processed, consumed, err
	}
	return processed, consumed, nil
}

// ReverseOrderCommissions 回滚指定订单关联的佣金记录（仅未提现部分）。
//
// 规则：
//   - 选取该订单下状态为 frozen/available 的佣金记录；
//   - 将其状态置为 reversed，并记录一条 adjust 流水，便于审计；
//   - 已提现（paid）的佣金不做自动回滚，需走人工财务调整。
func ReverseOrderCommissions(orderID uint, operatorID *uint, note string) (int, error) {
	if orderID == 0 {
		return 0, errors.New("orderID cannot be zero")
	}

	db, err := requireDB()
	if err != nil {
		return 0, err
	}

	var list []model.Commission
	if err := db.Where("order_id = ? AND status IN ?", orderID, []string{StatusFrozen, StatusAvailable}).
		Order("id ASC").
		Find(&list).Error; err != nil {
		return 0, err
	}
	if len(list) == 0 {
		return 0, nil
	}

	tx := db.Begin()
	processed := 0
	for _, cm := range list {
		// 再次限定状态，避免并发状态变化
		if err := tx.Model(&model.Commission{}).
			Where("id = ? AND status IN ?", cm.ID, []string{StatusFrozen, StatusAvailable}).
			Update("status", StatusReversed).Error; err != nil {
			tx.Rollback()
			return processed, err
		}

		adj := model.CommissionTransaction{
			CommissionID: cm.ID,
			Type:         "adjust",
			Amount:       cm.NetAmount,
			BalanceAfter: decimal.NewFromInt(0),
			Note:         note,
		}
		if operatorID != nil {
			adj.OperatorID = operatorID
		}
		if err := tx.Create(&adj).Error; err != nil {
			tx.Rollback()
			return processed, err
		}
		processed++
	}

	if err := tx.Commit().Error; err != nil {
		return processed, err
	}
	return processed, nil
}

func requireDB() (*gorm.DB, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database is not initialized")
	}
	return db, nil
}
