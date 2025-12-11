package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestCommissionReleaseRoute_Basic 确保手动佣金解冻路由可用且返回 JSON
func TestCommissionReleaseRoute_Basic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := NewCommissionAdminHandler()
	r.POST("/api/v1/admin/finance/commission/release", h.TriggerRelease)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/finance/commission/release", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 只要求返回是合法 JSON，具体业务状态由集成测试覆盖
	var body any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
}

// TestCommissionReverseOrderRoute_Basic 确保按订单回滚路由可用且返回 JSON
func TestCommissionReverseOrderRoute_Basic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h := NewCommissionAdminHandler()
	r.POST("/api/v1/admin/finance/commission/reverse-order", h.ReverseOrder)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/finance/commission/reverse-order", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 只要求返回是合法 JSON（错误也可以），具体业务逻辑由集成测试覆盖
	var body any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
}
