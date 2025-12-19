package handler

import (
	"tea-api/internal/config"
	"testing"
)

func TestCalcWithdrawalFee_Defaults(t *testing.T) {
	// Configure defaults: min 100, fixed 0, rate 30 bp (0.30%), cap 0
	config.Config.Finance.Withdrawal.MinAmountCents = 100
	config.Config.Finance.Withdrawal.FeeFixedCents = 0
	config.Config.Finance.Withdrawal.FeeRateBp = 30
	config.Config.Finance.Withdrawal.FeeMinCents = 100
	config.Config.Finance.Withdrawal.FeeCapCents = 0

	// 100000 cents => 0.30% = 300 cents
	fee := calcWithdrawalFee(100000)
	if fee != 300 {
		t.Fatalf("expected fee=300, got=%d", fee)
	}

	// 100 cents => rate fee=0 (rounded), min applies => 100
	fee2 := calcWithdrawalFee(100)
	if fee2 != 100 {
		t.Fatalf("expected fee=100, got=%d", fee2)
	}
}

func TestCalcWithdrawalFee_WithFixedAndCap(t *testing.T) {
	config.Config.Finance.Withdrawal.MinAmountCents = 100
	config.Config.Finance.Withdrawal.FeeFixedCents = 250
	config.Config.Finance.Withdrawal.FeeRateBp = 100 // 1.00%
	config.Config.Finance.Withdrawal.FeeMinCents = 100
	config.Config.Finance.Withdrawal.FeeCapCents = 400

	// amount 20000 => rate 200, fixed 250 wins => 250, cap 400 not hit
	if f := calcWithdrawalFee(20000); f != 250 {
		t.Fatalf("expected fee=250, got=%d", f)
	}
	// amount 500000 => rate 5000, cap to 400
	if f := calcWithdrawalFee(500000); f != 400 {
		t.Fatalf("expected fee=400 (cap), got=%d", f)
	}
}
