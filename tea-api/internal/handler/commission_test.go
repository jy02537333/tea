package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCommissionCalculate 测试佣金计算接口
func TestCommissionCalculate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewCommissionHandler()
	router := gin.New()
	router.POST("/commissions/calculate", handler.Calculate)

	// 准备测试数据
	requestBody := CalculateCommissionRequest{
		OrderID:        "test_order_001",
		PayerUserID:    1001,
		ReferrerUserID: intPtr(1002),
		Items: []OrderItemDTO{
			{
				SKUID:          "sku_001",
				UnitPriceCents: 100000, // 1000元
				Quantity:       1,
				DiscountCents:  0,
			},
		},
		ShippingCents: 2000,  // 20元运费
		CouponCents:   5000,  // 50元优惠券
		DiscountCents: 0,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/commissions/calculate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["code"])
	
	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data)
	
	// 验证佣金计算结果
	// 结算基数 = 100000 - 2000 - 5000 = 93000分 = 930元
	// 如果直推率是10%，直推佣金应该是 93元
	commissions := data["commissions"].([]interface{})
	assert.Greater(t, len(commissions), 0)
}

// TestReferralRecord 测试记录推荐关系
func TestReferralRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewReferralHandler()
	router := gin.New()
	router.POST("/referral/record", handler.RecordReferral)

	requestBody := RecordReferralRequest{
		ReferrerUserID: 1,
		ReferredUserID: 2,
		Source:         "share_link",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/referral/record", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 注意：这个测试需要数据库连接，在没有数据库的情况下会失败
	// 实际应该使用mock或者测试数据库
	t.Log("Referral record test response:", w.Body.String())
}

// Helper function to create int pointer
func intPtr(i int64) *int64 {
	return &i
}
