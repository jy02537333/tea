package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type unifiedOrderResp struct {
	Code int `json:"code"`
	Data struct {
		PaymentNo string `json:"payment_no"`
		Sign      string `json:"sign"`
	} `json:"data"`
}

type orderDetailResp struct {
	Code int `json:"code"`
	Data struct {
		Order struct {
			Status    int `json:"status"`
			PayStatus int `json:"pay_status"`
		} `json:"order"`
	} `json:"data"`
}

func Test_PaymentUnifiedOrderCallback(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	token := devLogin(t, ts, fmt.Sprintf("payment_user_%d", time.Now().UnixNano()))
	authHeader := "Bearer " + token

	catID := createCategory(t, ts, authHeader)
	productID := createProduct(t, ts, authHeader, catID)
	addCartItem(t, ts, authHeader, productID)
	orderID := createOrderFromCart(t, ts, authHeader)

	payRes := callUnifiedOrder(t, ts, authHeader, orderID)
	if payRes.Data.PaymentNo == "" {
		t.Fatalf("expected payment_no in unified order response: %+v", payRes)
	}
	if payRes.Data.Sign == "" {
		t.Fatalf("expected sign in unified order response: %+v", payRes)
	}

	triggerCallback(t, ts, payRes.Data.PaymentNo, payRes.Data.Sign)

	detail := fetchOrderDetail(t, ts, authHeader, orderID)
	if detail.Data.Order.PayStatus != 2 {
		t.Fatalf("order pay_status not updated, got %d", detail.Data.Order.PayStatus)
	}
	if detail.Data.Order.Status != 2 {
		t.Fatalf("order status not updated to paid, got %d", detail.Data.Order.Status)
	}
}

func createCategory(t *testing.T, ts *httptest.Server, auth string) uint {
	t.Helper()
	payload := map[string]any{"name": fmt.Sprintf("支付测试分类-%d", time.Now().UnixNano())}
	resp := authedJSONRequest(t, ts.URL+"/api/v1/categories", auth, payload)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create category status: %d", resp.StatusCode)
	}
	var data struct {
		Code int `json:"code"`
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("decode category resp: %v", err)
	}
	if data.Data.ID == 0 {
		t.Fatalf("category id is zero")
	}
	return data.Data.ID
}

func createProduct(t *testing.T, ts *httptest.Server, auth string, categoryID uint) uint {
	t.Helper()
	payload := map[string]any{
		"category_id": categoryID,
		"name":        fmt.Sprintf("支付测试商品-%d", time.Now().UnixNano()),
		"price":       "18.80",
		"stock":       5,
		"status":      1,
	}
	resp := authedJSONRequest(t, ts.URL+"/api/v1/products", auth, payload)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create product status: %d", resp.StatusCode)
	}
	var data struct {
		Code int `json:"code"`
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("decode product resp: %v", err)
	}
	if data.Data.ID == 0 {
		t.Fatalf("product id is zero")
	}
	return data.Data.ID
}

func addCartItem(t *testing.T, ts *httptest.Server, auth string, productID uint) {
	t.Helper()
	payload := map[string]any{
		"product_id": productID,
		"quantity":   1,
	}
	resp := authedJSONRequest(t, ts.URL+"/api/v1/cart", auth, payload)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("add cart status: %d", resp.StatusCode)
	}
}

func createOrderFromCart(t *testing.T, ts *httptest.Server, auth string) uint {
	t.Helper()
	payload := map[string]any{
		"delivery_type": 1,
		"remark":        "payment flow test",
	}
	resp := authedJSONRequest(t, ts.URL+"/api/v1/orders/from-cart", auth, payload)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create order status: %d", resp.StatusCode)
	}
	var data struct {
		Code int `json:"code"`
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("decode order resp: %v", err)
	}
	if data.Data.ID == 0 {
		t.Fatalf("order id is zero")
	}
	return data.Data.ID
}

func callUnifiedOrder(t *testing.T, ts *httptest.Server, auth string, orderID uint) unifiedOrderResp {
	t.Helper()
	payload := map[string]any{"order_id": orderID}
	resp := authedJSONRequest(t, ts.URL+"/api/v1/payments/unified-order", auth, payload)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unified order status: %d", resp.StatusCode)
	}
	var data unifiedOrderResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("decode unified order resp: %v", err)
	}
	if data.Code != 0 {
		t.Fatalf("unified order returned error code: %+v", data)
	}
	return data
}

func triggerCallback(t *testing.T, ts *httptest.Server, paymentNo, sign string) {
	t.Helper()
	payload := map[string]any{
		"app_id":         "tea-app-mock",
		"payment_no":     paymentNo,
		"transaction_id": fmt.Sprintf("mock_%s", paymentNo),
		"trade_state":    "SUCCESS",
		"paid_at":        time.Now().Format(time.RFC3339),
		"sign":           sign,
		"test_mode":      true,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(ts.URL+"/api/v1/payments/callback", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("callback request err: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("callback status: %d", resp.StatusCode)
	}
	var data struct {
		Code int `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("decode callback resp: %v", err)
	}
	if data.Code != 0 {
		t.Fatalf("callback returned error code: %+v", data)
	}
}

func fetchOrderDetail(t *testing.T, ts *httptest.Server, auth string, orderID uint) orderDetailResp {
	t.Helper()
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/orders/%d", ts.URL, orderID), nil)
	req.Header.Set("Authorization", auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("order detail request err: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("order detail status: %d", resp.StatusCode)
	}
	var detail orderDetailResp
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		t.Fatalf("decode order detail resp: %v", err)
	}
	return detail
}

func authedJSONRequest(t *testing.T, url, auth string, payload map[string]any) *http.Response {
	t.Helper()
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request %s err: %v", url, err)
	}
	return resp
}
