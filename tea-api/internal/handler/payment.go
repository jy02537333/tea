package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tea-api/internal/config"
	"tea-api/internal/service"
	"tea-api/pkg/response"
)

type PaymentHandler struct{ svc *service.PaymentService }

func NewPaymentHandler() *PaymentHandler { return &PaymentHandler{svc: service.NewPaymentService()} }

type createIntentReq struct {
	OrderID uint `json:"order_id" binding:"required"`
	Method  int  `json:"method"` // 1:微信 2:支付宝（模拟）
}

// CreateIntent 创建支付意图（需登录）
func (h *PaymentHandler) CreateIntent(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))
	var req createIntentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Method == 0 {
		req.Method = 1
	}
	pay, url, err := h.svc.CreateIntent(userID, req.OrderID, req.Method)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"payment_no": pay.PaymentNo, "amount": pay.Amount, "pay_url": url})
}

// MockCallback 模拟支付回调（仅 local/dev 生效）
func (h *PaymentHandler) MockCallback(c *gin.Context) {
	if env := config.Config.System.Env; env != "local" && env != "dev" {
		response.Forbidden(c, "仅开发环境可用")
		return
	}
	var req struct {
		PaymentNo string `json:"payment_no" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := h.svc.MockCallback(req.PaymentNo); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}
