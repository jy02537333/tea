package service

import (
	"encoding/json"
	"testing"
)

func TestBuildStoreWithdrawApplyRemark_JSON(t *testing.T) {
	s := buildStoreWithdrawApplyRemark(100000, 300, 99700)
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("remark not valid JSON: %v", err)
	}
	if m["phase"] != "freeze" {
		t.Fatalf("unexpected phase: %v", m["phase"])
	}
	if m["currency"] != "CNY" {
		t.Fatalf("unexpected currency: %v", m["currency"])
	}
	// numbers in JSON unmarshal to float64
	if int64(m["amount_cents"].(float64)) != 100000 {
		t.Fatalf("amount_cents mismatch: %v", m["amount_cents"])
	}
	if int64(m["fee_cents"].(float64)) != 300 {
		t.Fatalf("fee_cents mismatch: %v", m["fee_cents"])
	}
	if int64(m["net_cents"].(float64)) != 99700 {
		t.Fatalf("net_cents mismatch: %v", m["net_cents"])
	}
}
