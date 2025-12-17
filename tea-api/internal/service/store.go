package service

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/internal/service/commission"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

type StoreService struct{ db *gorm.DB }

func NewStoreService() *StoreService { return &StoreService{db: database.GetDB()} }

type StoreWalletSummary struct {
	StoreID        uint            `json:"store_id"`
	TotalPaid      decimal.Decimal `json:"total_paid"`
	TotalRefunded  decimal.Decimal `json:"total_refunded"`
	TotalWithdrawn decimal.Decimal `json:"total_withdrawn"`
	Available      decimal.Decimal `json:"available"`
}

// StoreFinanceTransaction 门店资金流水项
// 统一展示收款、退款、提现等资金变动记录
type StoreFinanceTransaction struct {
	ID        uint            `json:"id"`
	StoreID   uint            `json:"store_id"`
	Type      string          `json:"type"`      // payment/refund/withdraw
	Direction string          `json:"direction"` // in/out
	Amount    decimal.Decimal `json:"amount"`
	Fee       decimal.Decimal `json:"fee"`
	RelatedID uint            `json:"related_id"` // 关联订单ID或0
	RelatedNo string          `json:"related_no"` // 订单号/支付单号/退款单号/提现单号
	Method    int             `json:"method"`     // 支付方式（对支付/退款有效）
	Remark    string          `json:"remark"`
	CreatedAt time.Time       `json:"created_at"`
}

// ListStoreAccounts 列出门店收款账户
func (s *StoreService) ListStoreAccounts(storeID uint) ([]model.StoreBankAccount, error) {
	if storeID == 0 {
		return nil, errors.New("无效的门店ID")
	}
	// 校验门店存在
	if _, err := s.GetStore(storeID); err != nil {
		return nil, err
	}
	var list []model.StoreBankAccount
	if err := s.db.Where("store_id = ?", storeID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// CreateStoreAccount 为门店创建收款账户
func (s *StoreService) CreateStoreAccount(storeID uint, accountType, accountName, accountNo, bankName string, isDefault bool) (*model.StoreBankAccount, error) {
	if storeID == 0 {
		return nil, errors.New("无效的门店ID")
	}
	if accountName == "" || accountNo == "" {
		return nil, errors.New("账户名和账号必填")
	}

	// 校验门店存在
	if _, err := s.GetStore(storeID); err != nil {
		return nil, err
	}

	acc := &model.StoreBankAccount{
		StoreID:     storeID,
		AccountType: accountType,
		AccountName: accountName,
		AccountNo:   accountNo,
		BankName:    bankName,
		IsDefault:   isDefault,
	}

	return acc, s.db.Transaction(func(tx *gorm.DB) error {
		// 若设为默认，则先将该门店其他账户取消默认
		if isDefault {
			if err := tx.Model(&model.StoreBankAccount{}).
				Where("store_id = ? AND is_default = ?", storeID, true).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		// 若该门店当前无账户，则自动设为默认
		if !isDefault {
			var count int64
			if err := tx.Model(&model.StoreBankAccount{}).Where("store_id = ?", storeID).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				acc.IsDefault = true
			}
		}
		if err := tx.Create(acc).Error; err != nil {
			return err
		}
		return nil
	})
}

// UpdateStoreAccount 更新门店收款账户
func (s *StoreService) UpdateStoreAccount(storeID, accountID uint, accountType, accountName, accountNo, bankName *string, isDefault *bool) error {
	if storeID == 0 || accountID == 0 {
		return errors.New("无效的ID")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var acc model.StoreBankAccount
		if err := tx.Where("id = ? AND store_id = ?", accountID, storeID).First(&acc).Error; err != nil {
			return err
		}

		updates := map[string]any{"updated_at": time.Now()}
		if accountType != nil {
			updates["account_type"] = *accountType
		}
		if accountName != nil {
			updates["account_name"] = *accountName
		}
		if accountNo != nil {
			updates["account_no"] = *accountNo
		}
		if bankName != nil {
			updates["bank_name"] = *bankName
		}

		if isDefault != nil {
			if *isDefault {
				// 将该门店其他账户取消默认
				if err := tx.Model(&model.StoreBankAccount{}).
					Where("store_id = ? AND id <> ? AND is_default = ?", storeID, accountID, true).
					Update("is_default", false).Error; err != nil {
					return err
				}
				updates["is_default"] = true
			} else {
				updates["is_default"] = false
			}
		}

		if err := tx.Model(&model.StoreBankAccount{}).
			Where("id = ? AND store_id = ?", accountID, storeID).
			Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeleteStoreAccount 删除门店收款账户
func (s *StoreService) DeleteStoreAccount(storeID, accountID uint) error {
	if storeID == 0 || accountID == 0 {
		return errors.New("无效的ID")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var acc model.StoreBankAccount
		if err := tx.Where("id = ? AND store_id = ?", accountID, storeID).First(&acc).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.StoreBankAccount{}, acc.ID).Error; err != nil {
			return err
		}

		// 如果删除的是默认账户且仍有其他账户，则将最新的一条设为默认
		if acc.IsDefault {
			var latest model.StoreBankAccount
			if err := tx.Where("store_id = ?", storeID).
				Order("id DESC").
				First(&latest).Error; err == nil {
				if err := tx.Model(&model.StoreBankAccount{}).
					Where("id = ?", latest.ID).
					Update("is_default", true).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (s *StoreService) CreateStore(st *model.Store) error {
	if st.Name == "" {
		return errors.New("门店名称必填")
	}
	return s.db.Create(st).Error
}

func (s *StoreService) UpdateStore(id uint, updates map[string]any) error {
	return s.db.Model(&model.Store{}).Where("id = ?", id).Updates(updates).Error
}

func (s *StoreService) DeleteStore(id uint) error {
	return s.db.Delete(&model.Store{}, id).Error
}

func (s *StoreService) GetStore(id uint) (*model.Store, error) {
	var st model.Store
	if err := s.db.First(&st, id).Error; err != nil {
		return nil, err
	}
	return &st, nil
}

// GetStoreWalletSummary 计算门店的「预计可提现余额」
// 简化规则：
//
//	总收入 = 已支付订单的支付金额之和（按 store_id 聚合）
//	总退款 = 已成功退款金额之和
//	总提现 = 已申请且处理中/已完成的提现净额之和（actual_amount）
//	可用余额 = max(总收入 - 总退款 - 总提现, 0)
func (s *StoreService) GetStoreWalletSummary(storeID uint) (*StoreWalletSummary, error) {
	if storeID == 0 {
		return nil, errors.New("无效的门店ID")
	}

	// 确认门店存在
	if _, err := s.GetStore(storeID); err != nil {
		return nil, err
	}

	var paidRes struct{ Total decimal.Decimal }
	if err := s.db.Table("payments AS p").
		Joins("JOIN orders AS o ON o.id = p.order_id").
		Where("o.store_id = ? AND p.status = ?", storeID, 2).
		Select("COALESCE(SUM(p.amount), 0) AS total").
		Scan(&paidRes).Error; err != nil {
		return nil, err
	}

	var refundRes struct{ Total decimal.Decimal }
	if err := s.db.Table("refunds AS r").
		Joins("JOIN orders AS o ON o.id = r.order_id").
		Where("o.store_id = ? AND r.status = ?", storeID, 2).
		Select("COALESCE(SUM(r.refund_amount), 0) AS total").
		Scan(&refundRes).Error; err != nil {
		return nil, err
	}

	var wdRes struct{ Total decimal.Decimal }
	if err := s.db.Model(&model.WithdrawRecord{}).
		Where("store_id = ? AND status IN (?)", storeID, []int{model.WithdrawStatusProcessing, model.WithdrawStatusCompleted}).
		Select("COALESCE(SUM(actual_amount), 0) AS total").
		Scan(&wdRes).Error; err != nil {
		return nil, err
	}

	available := paidRes.Total.Sub(refundRes.Total).Sub(wdRes.Total)
	if available.IsNegative() {
		available = decimal.NewFromInt(0)
	}

	return &StoreWalletSummary{
		StoreID:        storeID,
		TotalPaid:      paidRes.Total,
		TotalRefunded:  refundRes.Total,
		TotalWithdrawn: wdRes.Total,
		Available:      available,
	}, nil
}

// ListStoreFinanceTransactions 门店资金流水
// 根据门店、时间范围与类型聚合支付、退款与提现记录
func (s *StoreService) ListStoreFinanceTransactions(storeID uint, start, end string, page, limit int, txType string) ([]StoreFinanceTransaction, int64, error) {
	if storeID == 0 {
		return nil, 0, errors.New("无效的门店ID")
	}

	// 校验门店是否存在
	if _, err := s.GetStore(storeID); err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}

	db := s.db
	includePayments := txType == "" || txType == "payment"
	includeRefunds := txType == "" || txType == "refund"
	includeWithdraws := txType == "" || txType == "withdraw"

	result := make([]StoreFinanceTransaction, 0)

	if includePayments {
		var pays []model.Payment
		pq := db.Model(&model.Payment{}).
			Joins("JOIN orders o ON o.id = payments.order_id").
			Where("o.store_id = ? AND payments.status = ?", storeID, 2)
		if start != "" {
			pq = pq.Where("payments.created_at >= ?", start)
		}
		if end != "" {
			pq = pq.Where("payments.created_at <= ?", end)
		}
		if err := pq.Find(&pays).Error; err != nil {
			return nil, 0, err
		}
		for _, p := range pays {
			result = append(result, StoreFinanceTransaction{
				ID:        p.ID,
				StoreID:   storeID,
				Type:      "payment",
				Direction: "in",
				Amount:    p.Amount,
				Fee:       decimal.NewFromInt(0),
				RelatedID: p.OrderID,
				RelatedNo: p.PaymentNo,
				Method:    p.PaymentMethod,
				Remark:    "订单收款",
				CreatedAt: p.CreatedAt,
			})
		}
	}

	if includeRefunds {
		var refs []model.Refund
		rq := db.Model(&model.Refund{}).
			Preload("Payment").
			Joins("JOIN orders o ON o.id = refunds.order_id").
			Where("o.store_id = ? AND refunds.status = ?", storeID, 2)
		if start != "" {
			rq = rq.Where("refunds.created_at >= ?", start)
		}
		if end != "" {
			rq = rq.Where("refunds.created_at <= ?", end)
		}
		if err := rq.Find(&refs).Error; err != nil {
			return nil, 0, err
		}
		for _, r := range refs {
			method := 0
			if r.Payment.ID != 0 {
				method = r.Payment.PaymentMethod
			}
			result = append(result, StoreFinanceTransaction{
				ID:        r.ID,
				StoreID:   storeID,
				Type:      "refund",
				Direction: "out",
				Amount:    r.RefundAmount,
				Fee:       decimal.NewFromInt(0),
				RelatedID: r.OrderID,
				RelatedNo: r.RefundNo,
				Method:    method,
				Remark:    r.RefundReason,
				CreatedAt: r.CreatedAt,
			})
		}
	}

	if includeWithdraws {
		var wds []model.WithdrawRecord
		wq := db.Model(&model.WithdrawRecord{}).
			Where("store_id = ? AND status IN (?)", storeID, []int{model.WithdrawStatusProcessing, model.WithdrawStatusCompleted})
		if start != "" {
			wq = wq.Where("created_at >= ?", start)
		}
		if end != "" {
			wq = wq.Where("created_at <= ?", end)
		}
		if err := wq.Find(&wds).Error; err != nil {
			return nil, 0, err
		}
		for _, w := range wds {
			result = append(result, StoreFinanceTransaction{
				ID:        w.ID,
				StoreID:   storeID,
				Type:      "withdraw",
				Direction: "out",
				Amount:    w.ActualAmount,
				Fee:       w.Fee,
				RelatedID: 0,
				RelatedNo: w.WithdrawNo,
				Method:    0,
				Remark:    w.Remark,
				CreatedAt: w.CreatedAt,
			})
		}
	}

	// 按时间倒序、ID倒序排序
	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedAt.Equal(result[j].CreatedAt) {
			return result[i].ID > result[j].ID
		}
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	total := int64(len(result))
	startIdx := (page - 1) * limit
	if startIdx >= len(result) {
		return []StoreFinanceTransaction{}, total, nil
	}
	endIdx := startIdx + limit
	if endIdx > len(result) {
		endIdx = len(result)
	}

	return result[startIdx:endIdx], total, nil
}

// ListStoreWithdraws 按门店维度分页查询提现记录，支持按状态过滤
func (s *StoreService) ListStoreWithdraws(storeID uint, page, limit int, status *int) ([]model.WithdrawRecord, int64, error) {
	if storeID == 0 {
		return nil, 0, errors.New("无效的门店ID")
	}
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}

	q := s.db.Model(&model.WithdrawRecord{}).Where("store_id = ?", storeID)
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.WithdrawRecord
	if err := q.Order("id DESC").Limit(limit).Offset((page - 1) * limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ApplyStoreWithdraw 为门店创建一条提现申请记录
func (s *StoreService) ApplyStoreWithdraw(storeID uint, amount decimal.Decimal, remark string, withdrawType int, operatorUserID uint) (*model.WithdrawRecord, error) {
	if storeID == 0 {
		return nil, errors.New("无效的门店ID")
	}
	if amount.LessThanOrEqual(decimal.NewFromInt(0)) {
		return nil, errors.New("提现金额必须大于0")
	}

	// 钱包余额校验
	summary, err := s.GetStoreWalletSummary(storeID)
	if err != nil {
		return nil, err
	}
	if amount.GreaterThan(summary.Available) {
		return nil, errors.New("提现金额超过可用余额")
	}

	// 读取系统配置中的最小提现金额和手续费率（若缺失则使用安全默认值）
	minAmt := decimal.NewFromFloat(100.0)
	feeRate := 0.006
	var cfg model.SystemConfig
	if err := s.db.Where("config_key = ? AND status = 1", "withdraw_min_amount").First(&cfg).Error; err == nil {
		if f, err2 := strconv.ParseFloat(cfg.ConfigValue, 64); err2 == nil && f > 0 {
			minAmt = decimal.NewFromFloat(f)
		}
	}
	if err := s.db.Where("config_key = ? AND status = 1", "withdraw_fee_rate").First(&cfg).Error; err == nil {
		if f, err2 := strconv.ParseFloat(cfg.ConfigValue, 64); err2 == nil && f >= 0 {
			feeRate = f
		}
	}
	if amount.LessThan(minAmt) {
		return nil, fmt.Errorf("单次最小提现金额为 %.2f 元", minAmt.InexactFloat64())
	}

	// 使用分为单位计算手续费与净额
	amountCents := amount.Mul(decimal.NewFromInt(100)).IntPart()
	netCents, feeCents := commission.ApproveWithdrawal(amountCents, feeRate)
	netAmt := decimal.NewFromInt(netCents).Div(decimal.NewFromInt(100))
	feeAmt := decimal.NewFromInt(feeCents).Div(decimal.NewFromInt(100))

	if netAmt.LessThanOrEqual(decimal.NewFromInt(0)) {
		return nil, errors.New("计算后的净额无效")
	}

	rec := &model.WithdrawRecord{
		UserID:       operatorUserID,
		StoreID:      storeID,
		WithdrawNo:   generateStoreWithdrawNo("SW"),
		Amount:       amount,
		Fee:          feeAmt,
		ActualAmount: netAmt,
		WithdrawType: withdrawType,
		Status:       model.WithdrawStatusPending,
		Remark:       remark,
	}
	if err := s.db.Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

// ListStores 支持按经纬度计算距离并排序（在内存计算，便于 SQLite 测试环境）
func (s *StoreService) ListStores(page, limit int, status *int, lat, lng *float64) ([]map[string]any, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	q := s.db.Model(&model.Store{})
	if status != nil {
		q = q.Where("status = ?", *status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.Store
	if err := q.Order("id desc").Limit(limit).Offset((page - 1) * limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	res := make([]map[string]any, 0, len(list))
	for _, st := range list {
		item := map[string]any{
			"id":             st.ID,
			"name":           st.Name,
			"address":        st.Address,
			"phone":          st.Phone,
			"latitude":       st.Latitude,
			"longitude":      st.Longitude,
			"business_hours": st.BusinessHours,
			"images":         st.Images,
			"status":         st.Status,
		}
		if lat != nil && lng != nil && st.Latitude != 0 && st.Longitude != 0 {
			item["distance_km"] = haversine(*lat, *lng, st.Latitude, st.Longitude)
		}
		res = append(res, item)
	}

	// 简化：分页后不再排序；若需要严格距离排序，可在内存排序，但需取全量或扩大分页
	return res, total, nil
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // km
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return math.Round((R*c)*1000) / 1000 // 保留3位小数
}

func toRad(d float64) float64 { return d * math.Pi / 180.0 }

// generateStoreWithdrawNo 生成门店提现单号
func generateStoreWithdrawNo(prefix string) string {
	ts := time.Now().Format("20060102150405")
	uid := utils.GenerateUID()
	if len(uid) > 6 {
		uid = uid[:6]
	}
	return fmt.Sprintf("%s%s%s", prefix, ts, uid)
}
