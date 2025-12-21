package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/database"
	"tea-api/pkg/response"
)

type CouponHandler struct{ svc *service.CouponService }

func NewCouponHandler() *CouponHandler { return &CouponHandler{svc: service.NewCouponService()} }

type couponRequest struct {
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

func parseCouponPayload(c *gin.Context) (*model.Coupon, error) {
	var req couponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	parseDec := func(s string) (decimal.Decimal, error) {
		if s == "" {
			return decimal.Zero, nil
		}
		return decimal.NewFromString(s)
	}
	amount, err := parseDec(req.Amount)
	if err != nil {
		return nil, errors.New("金额格式不正确")
	}
	discount, err := parseDec(req.Discount)
	if err != nil {
		return nil, errors.New("折扣格式不正确")
	}
	minAmt, err := parseDec(req.MinAmount)
	if err != nil {
		return nil, errors.New("门槛金额格式不正确")
	}
	st, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, errors.New("开始时间格式不正确")
	}
	et, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, errors.New("结束时间格式不正确")
	}
	return &model.Coupon{
		Name: req.Name, Type: req.Type,
		Amount: amount, Discount: discount, MinAmount: minAmt,
		TotalCount: req.TotalCount, Status: req.Status,
		StartTime: st, EndTime: et, Description: req.Description,
	}, nil
}

// CreateCoupon 管理端创建优惠券（简化：只需登录）
func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	coupon, err := parseCouponPayload(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
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

// ListStoreCoupons 门店维度列出优惠券
func (h *CouponHandler) ListStoreCoupons(c *gin.Context) {
	status := 0
	if v := c.Query("status"); v != "" {
		status = atoi(v)
	}
	storeID := uint(atoi(c.Param("id")))
	if storeID == 0 {
		response.BadRequest(c, "无效的门店ID")
		return
	}
	list, err := h.svc.ListStoreCoupons(storeID, status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, list)
}

// CreateStoreCoupon 门店创建本店优惠券
func (h *CouponHandler) CreateStoreCoupon(c *gin.Context) {
	storeID := uint(atoi(c.Param("id")))
	if storeID == 0 {
		response.BadRequest(c, "无效的门店ID")
		return
	}
	coupon, err := parseCouponPayload(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CreateStoreCoupon(storeID, coupon); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, coupon)
}

// UpdateStoreCoupon 门店编辑本店优惠券
func (h *CouponHandler) UpdateStoreCoupon(c *gin.Context) {
	storeID := uint(atoi(c.Param("id")))
	if storeID == 0 {
		response.BadRequest(c, "无效的门店ID")
		return
	}
	couponID := uint(atoi(c.Param("couponId")))
	if couponID == 0 {
		response.BadRequest(c, "无效的优惠券ID")
		return
	}
	upd, err := parseCouponPayload(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	res, err := h.svc.UpdateStoreCoupon(storeID, couponID, upd)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, res)
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

// ListCouponTemplates GET /api/v1/coupons/templates
// 简化：直接读取 coupons_templates 表，返回全部或分页（此处返回全部）
func ListCouponTemplates(c *gin.Context) {
	type tplRow struct {
		ID             uint   `json:"id"`
		Name           string `json:"name"`
		Type           string `json:"type"`
		Value          int64  `json:"value"`
		MinOrderAmount int64  `json:"min_order_amount"`
		TotalQuantity  int    `json:"total_quantity"`
		ValidFrom      string `json:"valid_from"`
		ValidTo        string `json:"valid_to"`
	}
	var list []tplRow
	if err := database.GetDB().Table("coupons_templates").
		Select("id,name,type,value,min_order_amount,total_quantity,DATE_FORMAT(valid_from,'%Y-%m-%dT%H:%i:%sZ') as valid_from, DATE_FORMAT(valid_to,'%Y-%m-%dT%H:%i:%sZ') as valid_to").
		Order("id desc").Find(&list).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "查询模板失败")
		return
	}
	response.Success(c, list)
}

// ClaimCouponFromTemplate POST /api/v1/coupons/claim
// Body: {template_id}
func ClaimCouponFromTemplate(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未获取到用户身份")
		return
	}
	var req struct {
		TemplateID uint `json:"template_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.TemplateID == 0 {
		response.BadRequest(c, "参数错误")
		return
	}
	// 生成简单券码
	code := fmt.Sprintf("U%d-T%d-%d", uid, req.TemplateID, time.Now().Unix())
	// 读取模板有效期作为券的过期时间（简化处理）
	type tplRow struct{ ValidTo time.Time }
	var tpl tplRow
	if err := database.GetDB().Table("coupons_templates").Select("valid_to").Where("id = ?", req.TemplateID).Take(&tpl).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "模板不存在")
		return
	}
	// 创建用户券
	tx := database.GetDB().Exec("INSERT INTO coupons (template_id, user_id, code, status, expires_at, claimed_at) VALUES (?,?,?,?,?,NOW())",
		req.TemplateID, uid, code, "unused", tpl.ValidTo)
	if tx.Error != nil {
		response.Error(c, http.StatusInternalServerError, "领取失败")
		return
	}
	response.Success(c, gin.H{"coupon_code": code, "expires_at": tpl.ValidTo.Format(time.RFC3339)})
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
