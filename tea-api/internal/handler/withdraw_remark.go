package handler

import (
	"encoding/json"
)

// buildFreezeRemark 构造用户申请冻结阶段的钱包流水备注(JSON)
func buildFreezeRemark(amountCents, feeCents, netCents int64) string {
	m := map[string]any{
		"phase":        "freeze",
		"amount_cents": amountCents,
		"fee_cents":    feeCents,
		"net_cents":    netCents,
		"currency":     "CNY",
	}
	b, _ := json.Marshal(m)
	return string(b)
}

// buildPaidRemark 构造管理员完成(paid)阶段的钱包流水备注(JSON)
func buildPaidRemark(withdrawNo string, amountCents, feeCents, netCents int64) string {
	m := map[string]any{
		"phase":        "paid",
		"withdraw_no":  withdrawNo,
		"amount_cents": amountCents,
		"fee_cents":    feeCents,
		"net_cents":    netCents,
		"currency":     "CNY",
	}
	b, _ := json.Marshal(m)
	return string(b)
}

// buildRejectUnfreezeRemark 构造管理员拒绝(rejected_unfreeze)阶段的钱包流水备注(JSON)
func buildRejectUnfreezeRemark(withdrawNo string, amountCents, feeCents, netCents int64) string {
	m := map[string]any{
		"phase":        "rejected_unfreeze",
		"withdraw_no":  withdrawNo,
		"amount_cents": amountCents,
		"fee_cents":    feeCents,
		"net_cents":    netCents,
		"currency":     "CNY",
	}
	b, _ := json.Marshal(m)
	return string(b)
}
