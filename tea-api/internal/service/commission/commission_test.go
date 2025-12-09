package commission

import (
	"testing"
	"time"
)

func TestCalculateDirectCommission(t *testing.T) {
	basis := int64(93000) // 930.00 元
	rate := 0.30
	want := int64(27900) // 279.00 元
	got := CalculateDirectCommission(basis, rate)
	if got != want {
		t.Fatalf("direct commission: want %d got %d", want, got)
	}
}

func TestCalculateIndirectCommission(t *testing.T) {
	directGross := int64(27900)
	indirectRate := 0.10
	want := int64(2790)
	got := CalculateIndirectCommission(directGross, indirectRate)
	if got != want {
		t.Fatalf("indirect commission: want %d got %d", want, got)
	}
}

func TestBuildCommissionRecords(t *testing.T) {
	order := Order{
		ID:             1001,
		UserID:         2001,
		TotalAmount:    100000, // 1000.00
		ShippingAmount: 2000,   // 20.00
		CouponAmount:   5000,   // 50.00
		DiscountAmount: 0,
	}
	recs := BuildCommissionRecords(order, 3001, 0.30, 0.10, 7)
	if len(recs) != 2 {
		t.Fatalf("expected 2 commission records (direct+indirect), got %d", len(recs))
	}
	direct := recs[0]
	if direct.UserID != 3001 || direct.CommissionType != "direct" {
		t.Fatalf("direct record mismatch: %+v", direct)
	}
	if direct.GrossAmount != 27900 {
		t.Fatalf("direct gross wrong: want 27900 got %d", direct.GrossAmount)
	}
	indirect := recs[1]
	if indirect.CommissionType != "indirect" {
		t.Fatalf("indirect record mismatch: %+v", indirect)
	}
	if indirect.GrossAmount != 2790 {
		t.Fatalf("indirect gross wrong: want 2790 got %d", indirect.GrossAmount)
	}
	// available_at 应约等于 now + 7 天（允许少许时间差）
	if time.Until(direct.AvailableAt) < 6*24*time.Hour {
		t.Fatalf("direct AvailableAt too soon: %v", direct.AvailableAt)
	}
}

func TestApproveWithdrawal(t *testing.T) {
	amount := int64(100000) // 1000.00 元
	feeRate := 0.01
	net, fee := ApproveWithdrawal(amount, feeRate)
	if fee != 1000 {
		t.Fatalf("expected fee 1000 got %d", fee)
	}
	if net != 99000 {
		t.Fatalf("expected net 99000 got %d", net)
	}
}
