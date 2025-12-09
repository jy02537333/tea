package commission

import (
	"math"
	"time"
)

// 简化的领域模型（用于计算逻辑）
type Order struct {
	ID             int64
	UserID         int64
	TotalAmount    int64 // 订单总金额（分）
	ShippingAmount int64 // 运费（分）
	CouponAmount   int64 // 优惠券抵扣（分）
	DiscountAmount int64 // 其他折扣（分）
	Items          []OrderItem
}

type OrderItem struct {
	ID        int64
	SKU       string
	Quantity  int
	UnitPrice int64 // 单价（分）
}

type CommissionRecord struct {
	UserID           int64
	OrderID          int64
	OrderItemID      int64
	CommissionType   string  // direct|indirect|upgrade
	SourceUserID     int64   // 触发该佣金的用户（如下单用户）
	Rate             float64 // 比例，如 0.30
	CalculationBasis int64   // 计算基数（分）
	GrossAmount      int64   // 毛佣金（分）
	Fee              int64   // 手续费（分）
	NetAmount        int64   // 净佣金（分）
	AvailableAt      time.Time
}

// CalculateDirectCommission 按基数和比例计算毛佣金（向下取整）
func CalculateDirectCommission(basis int64, rate float64) int64 {
	raw := float64(basis) * rate
	return int64(math.Floor(raw))
}

// CalculateIndirectCommission: 间接佣金通常是直推佣金的某个比例（文档中为 10% 的团队管理奖）
func CalculateIndirectCommission(directGross int64, indirectRate float64) int64 {
	raw := float64(directGross) * indirectRate
	return int64(math.Floor(raw))
}

// BuildCommissionRecords 根据订单与传入的分佣配置生成佣金记录（内存计算）
// 参数：directRate 为直推比例（对推荐人），indirectRate 为间接比例（对上级对应的团队奖）
// 返回：生成的 CommissionRecord 列表（包含 direct + indirect）
func BuildCommissionRecords(order Order, directReferrerID int64, directRate float64, indirectRate float64, holdPeriodDays int) []CommissionRecord {
	// 计算结算基数：订单应付项减去运费/优惠/折扣（按需求：商品售价 - 快递费用 - 营销优惠费用 - 折扣费用）
	calculationBasis := order.TotalAmount - order.ShippingAmount - order.CouponAmount - order.DiscountAmount
	if calculationBasis < 0 {
		calculationBasis = 0
	}

	// 直接佣金
	directGross := CalculateDirectCommission(calculationBasis, directRate)
	direct := CommissionRecord{
		UserID:           directReferrerID,
		OrderID:          order.ID,
		CommissionType:   "direct",
		SourceUserID:     order.UserID,
		Rate:             directRate,
		CalculationBasis: calculationBasis,
		GrossAmount:      directGross,
		Fee:              0,
		NetAmount:        directGross,
		AvailableAt:      time.Now().Add(time.Duration(holdPeriodDays) * 24 * time.Hour),
	}

	// 间接佣金（对直接推荐人的上级）
	indirectGross := CalculateIndirectCommission(directGross, indirectRate)
	var indirect CommissionRecord
	if indirectGross > 0 {
		// 在此 stub 中我们不传入上级 ID（实际应通过 referrals_closure 查找祖先），用 0 占位
		indirect = CommissionRecord{
			UserID:           0,
			OrderID:          order.ID,
			CommissionType:   "indirect",
			SourceUserID:     order.UserID,
			Rate:             indirectRate,
			CalculationBasis: directGross,
			GrossAmount:      indirectGross,
			Fee:              0,
			NetAmount:        indirectGross,
			AvailableAt:      time.Now().Add(time.Duration(holdPeriodDays) * 24 * time.Hour),
		}
	}

	out := []CommissionRecord{direct}
	if indirectGross > 0 {
		out = append(out, indirect)
	}
	return out
}
