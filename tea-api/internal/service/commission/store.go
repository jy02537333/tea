package commission

import (
	"fmt"

	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

// SaveCommissionRecords 将内存计算的 CommissionRecord 列表持久化到数据库（事务）
func SaveCommissionRecords(records []CommissionRecord) error {
	if len(records) == 0 {
		return nil
	}
	db := database.GetDB()
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, r := range records {
		cm := model.Commission{
			UserID:         uint(r.UserID),
			CommissionType: r.CommissionType,
			Status:         StatusFrozen,
		}

		if r.OrderID != 0 {
			oid := uint(r.OrderID)
			cm.OrderID = &oid
		}
		if r.OrderItemID != 0 {
			oi := uint(r.OrderItemID)
			cm.OrderItemID = &oi
		}
		if r.SourceUserID != 0 {
			su := uint(r.SourceUserID)
			cm.SourceUserID = &su
		}

		// 货币从分转换为元（decimal）
		cm.CalculationBasis = decimal.NewFromInt(r.CalculationBasis).Div(decimal.NewFromInt(100))
		cm.GrossAmount = decimal.NewFromInt(r.GrossAmount).Div(decimal.NewFromInt(100))
		cm.Fee = decimal.NewFromInt(r.Fee).Div(decimal.NewFromInt(100))
		cm.NetAmount = decimal.NewFromInt(r.NetAmount).Div(decimal.NewFromInt(100))
		cm.Rate = decimal.NewFromFloat(r.Rate)
		if !r.AvailableAt.IsZero() {
			t := r.AvailableAt
			cm.AvailableAt = &t
		}

		if err := tx.Create(&cm).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("create commission failed: %w", err)
		}

		// 记录一条 commission transaction 的初始流水（创建/冻结）
		cTx := model.CommissionTransaction{
			CommissionID: cm.ID,
			Type:         "freeze",
			Amount:       cm.GrossAmount,
			BalanceAfter: cm.GrossAmount,
		}
		if err := tx.Create(&cTx).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("create commission tx failed: %w", err)
		}
	}

	return tx.Commit().Error
}
