package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/response"
)

type CouponHandler struct{ svc *service.CouponService }

func NewCouponHandler() *CouponHandler { return &CouponHandler{svc: service.NewCouponService()} }

// CreateCoupon 管理端创建优惠券（简化：只需登录）
func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		Type        int    `json:"type"`
		Amount      string `json:"amount"`
		Discount    string `json:"discount"`
		MinAmount   string `json:"min_amount"`
		TotalCount  int    `json:"total_count"`
		Status      int    `json:"status"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	parseDec := func(s string) (decimal.Decimal, error) {
		if s == "" {
			return decimal.Zero, nil
		}
		return decimal.NewFromString(s)
	}
	amount, err := parseDec(req.Amount)
	if err != nil {
		response.BadRequest(c, "金额格式不正确")
		return
	}
	discount, err := parseDec(req.Discount)
	if err != nil {
		response.BadRequest(c, "折扣格式不正确")
		return
	}
	minAmt, err := parseDec(req.MinAmount)
	if err != nil {
		response.BadRequest(c, "门槛金额格式不正确")
		return
	}
	st, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		response.BadRequest(c, "开始时间格式不正确")
		return
	}
	et, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		response.BadRequest(c, "结束时间格式不正确")
		return
	}
	coupon := &model.Coupon{
		Name: req.Name, Type: req.Type,
		Amount: amount, Discount: discount, MinAmount: minAmt,
		TotalCount: req.TotalCount, Status: req.Status,
		StartTime: st, EndTime: et, Description: req.Description,
	}
	if err := h.svc.CreateCoupon(coupon); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, coupon)
}

// ListCoupons 列出优惠券（简化：仅状态筛选）
func (h *CouponHandler) ListCoupons(c *gin.Context) {
	status := 0
	if v := c.Query("status"); v != "" {
		status = atoi(v)
	}
	list, err := h.svc.ListCoupons(status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, list)
}

// Grant 给用户发券
func (h *CouponHandler) Grant(c *gin.Context) {
	var req struct {
		CouponID uint `json:"coupon_id"`
		UserID   uint `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	// 兼容测试/调用方未传或传0的场景：默认对当前登录用户发券
	if req.UserID == 0 {
		if v, ok := c.Get("user_id"); ok {
			if uid, ok2 := v.(uint); ok2 {
				req.UserID = uid
			}
		}
	}
	uc, err := h.svc.GrantCouponToUser(req.CouponID, req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, uc)
}

// ListMyCoupons 当前用户可用优惠券
func (h *CouponHandler) ListMyCoupons(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))
	list, err := h.svc.ListUserAvailableCoupons(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, list)
}

func atoi(s string) int {
	v := 0
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch < '0' || ch > '9' {
			return v
		}
		v = v*10 + int(ch-'0')
	}
	return v
}
