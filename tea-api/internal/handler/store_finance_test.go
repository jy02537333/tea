package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestStoreWalletRoute_Basic ensures the /stores/:id/wallet route is wired and parses ID correctly.
func TestStoreWalletRoute_Basic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 使用真实的 StoreHandler，但不依赖具体 DB 数据，只验证路由/参数与 4xx/5xx 之外的行为
	h := NewStoreHandler()
	r.GET("/api/v1/stores/:id/wallet", h.Wallet)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stores/0/wallet", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		// 当 id=0 时，我们期望是 400，而不是成功
		t.Fatalf("expected non-200 status for invalid id, got %d", w.Code)
	}

	// 简单检查返回是 JSON
	var body any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
}

// TestStoreAccountsRoutes_Basic 确认门店收款账户路由已注册且基础返回结构正确
func TestStoreAccountsRoutes_Basic(t *testing.T) {
	r := gin.Default()
	h := NewStoreHandler()

	// 简单的假登录中间件，直接放行
	auth := func(c *gin.Context) { c.Next() }

	// 注册账户管理相关路由
	r.GET("/api/v1/stores/:id/accounts", auth, h.ListAccounts)
	r.POST("/api/v1/stores/:id/accounts", auth, h.CreateAccount)
	r.PUT("/api/v1/stores/:id/accounts/:accountId", auth, h.UpdateAccount)
	r.DELETE("/api/v1/stores/:id/accounts/:accountId", auth, h.DeleteAccount)

	// 使用非法ID，期望返回 4xx 或 5xx，而不是 200 空成功
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stores/0/accounts", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code == http.StatusOK {
		t.Fatalf("expected non-200 status for invalid store id, got %d", resp.Code)
	}
	// 响应应该是 JSON
	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
}

// TestStoreFinanceTransactionsRoute_Basic ensures the /stores/:id/finance/transactions route is wired.
func TestStoreFinanceTransactionsRoute_Basic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := NewStoreHandler()
	r.GET("/api/v1/stores/:id/finance/transactions", h.FinanceTransactions)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stores/0/finance/transactions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		// 当 id=0 时，我们期望是 400，而不是成功
		t.Fatalf("expected non-200 status for invalid id, got %d", w.Code)
	}

	var body any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
}

// TestStoreFinanceTransactionsExportRoute_Basic ensures the export route is wired.
func TestStoreFinanceTransactionsExportRoute_Basic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := NewStoreHandler()
	r.GET("/api/v1/stores/:id/finance/transactions/export", h.ExportFinanceTransactions)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stores/0/finance/transactions/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		// 当 id=0 时，我们期望是 400，而不是成功
		t.Fatalf("expected non-200 status for invalid id, got %d", w.Code)
	}

	// 响应应为 JSON 错误结构
	var body any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
}

// TestStoreFinanceTransactions_InvalidIDMessage checks error payload when id is invalid.
func TestStoreFinanceTransactions_InvalidIDMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := NewStoreHandler()
	r.GET("/api/v1/stores/:id/finance/transactions", h.FinanceTransactions)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stores/0/finance/transactions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 for invalid id, got %d", w.Code)
	}
	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if body.Message == "" {
		t.Fatalf("expected message, got empty")
	}
}

// TestStoreFinanceTransactionsExport_InvalidIDMessage checks export route error payload for invalid id.
func TestStoreFinanceTransactionsExport_InvalidIDMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := NewStoreHandler()
	r.GET("/api/v1/stores/:id/finance/transactions/export", h.ExportFinanceTransactions)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stores/0/finance/transactions/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 for invalid id, got %d", w.Code)
	}
	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if body.Message == "" {
		t.Fatalf("expected message, got empty")
	}
}
