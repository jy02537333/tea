package commission

import (
	"math"
)

// ApproveWithdrawal 计算提现手续费并返回净额（分）和手续费（分）
// amountCents: 申请提现金额（分）
// feeRate: 手续费率（例如 0.01 表示 1%）
func ApproveWithdrawal(amountCents int64, feeRate float64) (netCents int64, feeCents int64) {
	if amountCents <= 0 {
		return 0, 0
	}
	rawFee := float64(amountCents) * feeRate
	feeCents = int64(math.Floor(rawFee))
	netCents = amountCents - feeCents
	if netCents < 0 {
		netCents = 0
	}
	return netCents, feeCents
}
