package handler

import (
	"encoding/json"
	"testing"
)

func TestBuildFreezeRemark_JSON(t *testing.T) {
	s := buildFreezeRemark(100000, 300, 99700)
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("freeze remark not JSON: %v", err)
	}
	if m["phase"] != "freeze" {
		t.Fatalf("phase: %v", m["phase"])
	}
	if m["currency"] != "CNY" {
		t.Fatalf("currency: %v", m["currency"])
	}
	if v, ok := m["amount_cents"].(float64); !ok || int64(v) != 100000 {
		t.Fatalf("amount_cents: %v", m["amount_cents"])
	}
	if v, ok := m["fee_cents"].(float64); !ok || int64(v) != 300 {
		t.Fatalf("fee_cents: %v", m["fee_cents"])
	}
	if v, ok := m["net_cents"].(float64); !ok || int64(v) != 99700 {
		t.Fatalf("net_cents: %v", m["net_cents"])
	}
}

func TestBuildPaidRemark_JSON(t *testing.T) {
	s := buildPaidRemark("WD_NO_1", 100000, 300, 99700)
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("paid remark not JSON: %v", err)
	}
	if m["phase"] != "paid" {
		t.Fatalf("phase: %v", m["phase"])
	}
	if m["withdraw_no"] != "WD_NO_1" {
		t.Fatalf("withdraw_no: %v", m["withdraw_no"])
	}
}

func TestBuildRejectUnfreezeRemark_JSON(t *testing.T) {
	s := buildRejectUnfreezeRemark("WD_NO_2", 50000, 150, 49850)
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("reject remark not JSON: %v", err)
	}
	if m["phase"] != "rejected_unfreeze" {
		t.Fatalf("phase: %v", m["phase"])
	}
	if m["withdraw_no"] != "WD_NO_2" {
		t.Fatalf("withdraw_no: %v", m["withdraw_no"])
	}
}
