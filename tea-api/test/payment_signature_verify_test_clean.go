package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func Test_PaymentCallback_Signature_Failure(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	payload := map[string]any{
		"app_id":         "tea-app-mock",
		"payment_no":     fmt.Sprintf("PTEST_%d", time.Now().UnixNano()),
		"transaction_id": "txn_wrong",
		"trade_state":    "SUCCESS",
		"paid_at":        time.Now().Format(time.RFC3339),
		"sign":           "WRONGSIGN",
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(ts.URL+"/api/v1/payments/callback", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("callback request err: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected non-200 on signature failure, got %d", resp.StatusCode)
	}
}

func Test_PaymentCallback_Compat_Success(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	token := devLogin(t, ts, fmt.Sprintf("payment_user_%d", time.Now().UnixNano()))
	authHeader := "Bearer " + token

	catID := createCategory(t, ts, authHeader)
	productID := createProduct(t, ts, authHeader, catID)
	addCartItem(t, ts, authHeader, productID)
	orderID := createOrderFromCart(t, ts, authHeader)

	uo := callUnifiedOrder(t, ts, authHeader, orderID)
	if uo.Data.PaymentNo == "" {
		t.Fatalf("payment_no empty")
	}

	payload := map[string]any{
		"app_id":         "tea-app-mock",
		"payment_no":     uo.Data.PaymentNo,
		"transaction_id": fmt.Sprintf("mock_%s", uo.Data.PaymentNo),
		"trade_state":    "SUCCESS",
		"paid_at":        time.Now().Format(time.RFC3339),
		"sign":           uo.Data.Sign,
		"test_mode":      true,
	}
	b, _ := json.Marshal(payload)
	resp, err := http.Post(ts.URL+"/api/v1/payments/callback", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("callback request err: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("callback status: %d", resp.StatusCode)
	}
}
