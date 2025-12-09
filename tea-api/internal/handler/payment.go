package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

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

type unifiedOrderReq struct {
	OrderID   uint   `json:"order_id" binding:"required"`
	Method    int    `json:"method"`
	TradeType string `json:"trade_type"`
	NotifyURL string `json:"notify_url"`
}

type paymentCallbackReq struct {
	AppID         string `json:"app_id"`
	PaymentNo     string `json:"payment_no" binding:"required"`
	TransactionID string `json:"transaction_id"`
	TradeState    string `json:"trade_state" binding:"required"`
	PaidAt        string `json:"paid_at"`
	Sign          string `json:"sign"`
	TestMode      bool   `json:"test_mode"`
}

// UnifiedOrder 模拟微信统一下单接口
func (h *PaymentHandler) UnifiedOrder(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))
	var req unifiedOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Method == 0 {
		req.Method = 1
	}
	res, err := h.svc.UnifiedOrder(userID, req.OrderID, req.Method)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{
		"payment_no": res.PaymentNo,
		"order_id":   res.OrderID,
		"amount":     res.Amount,
		"prepay_id":  res.PrepayID,
		"nonce_str":  res.NonceStr,
		"timestamp":  res.Timestamp,
		"package":    res.Package,
		"sign":       res.Sign,
		"pay_url":    res.PayURL,
	})
}

// Callback 处理第三方支付回调
func (h *PaymentHandler) Callback(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "读取回调失败")
		return
	}
	var req paymentCallbackReq
	if err := json.Unmarshal(body, &req); err != nil {
		response.BadRequest(c, "回调报文格式错误")
		return
	}
	var paidAt *time.Time
	if req.PaidAt != "" {
		if t, err := time.Parse(time.RFC3339, req.PaidAt); err == nil {
			paidAt = &t
		}
	}
	skipVerify := false
	if req.TestMode {
		if env := config.Config.System.Env; env == "local" || env == "dev" {
			skipVerify = true
		} else {
			response.Forbidden(c, "test_mode 仅在本地/开发环境允许")
			return
		}
	}
	payload := service.PaymentCallbackPayload{
		PaymentNo:     req.PaymentNo,
		TransactionID: req.TransactionID,
		TradeState:    req.TradeState,
		Sign:          req.Sign,
		PaidAt:        paidAt,
		RawBody:       string(body),
		TestMode:      req.TestMode,
		SkipVerify:    skipVerify,
	}
	if err := h.svc.HandleCallback(payload); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
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
