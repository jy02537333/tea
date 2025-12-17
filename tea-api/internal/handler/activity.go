package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/response"
)

type ActivityHandler struct{ svc *service.ActivityService }

func NewActivityHandler() *ActivityHandler {
	return &ActivityHandler{svc: service.NewActivityService()}
}

type activityRequest struct {
	Name        string `json:"name"`
	Type        int    `json:"type"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Rules       string `json:"rules"`
	Status      int    `json:"status"`
	Priority    int    `json:"priority"`
	Description string `json:"description"`
}

func parseActivityPayload(c *gin.Context) (*model.Activity, error) {
	var req activityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	st, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, errors.New("开始时间格式不正确")
	}
	et, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, errors.New("结束时间格式不正确")
	}
	return &model.Activity{
		Name:        req.Name,
		Type:        req.Type,
		StartTime:   st,
		EndTime:     et,
		Rules:       req.Rules,
		Status:      req.Status,
		Priority:    req.Priority,
		Description: req.Description,
	}, nil
}

// ListActivities 用户侧：按门店列出当前可报名活动
// GET /api/v1/activities?store_id=
func (h *ActivityHandler) ListActivities(c *gin.Context) {
	storeID := uint(atoi(c.Query("store_id")))
	if storeID == 0 {
		response.BadRequest(c, "缺少或无效的门店ID")
		return
	}
	list, err := h.svc.ListUserActivities(storeID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, list)
}

// ListStoreActivities 门店维度列出活动
func (h *ActivityHandler) ListStoreActivities(c *gin.Context) {
	storeID := uint(atoi(c.Param("id")))
	if storeID == 0 {
		response.BadRequest(c, "无效的门店ID")
		return
	}
	status := 0
	if v := c.Query("status"); v != "" {
		status = atoi(v)
	}
	list, err := h.svc.ListStoreActivities(storeID, status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, list)
}

// CreateStoreActivity 门店创建本店活动
func (h *ActivityHandler) CreateStoreActivity(c *gin.Context) {
	storeID := uint(atoi(c.Param("id")))
	if storeID == 0 {
		response.BadRequest(c, "无效的门店ID")
		return
	}
	act, err := parseActivityPayload(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CreateStoreActivity(storeID, act); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, act)
}

// UpdateStoreActivity 门店编辑本店活动
func (h *ActivityHandler) UpdateStoreActivity(c *gin.Context) {
	storeID := uint(atoi(c.Param("id")))
	if storeID == 0 {
		response.BadRequest(c, "无效的门店ID")
		return
	}
	activityID := uint(atoi(c.Param("activityId")))
	if activityID == 0 {
		response.BadRequest(c, "无效的活动ID")
		return
	}
	upd, err := parseActivityPayload(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	res, err := h.svc.UpdateStoreActivity(storeID, activityID, upd)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, res)
}

// ListActivityRegistrations 活动报名列表（门店后台视角）
// GET /api/v1/stores/:id/activities/:activityId/registrations?page=&limit=&status=
func (h *ActivityHandler) ListActivityRegistrations(c *gin.Context) {
	storeID := uint(atoi(c.Param("id")))
	if storeID == 0 {
		response.BadRequest(c, "无效的门店ID")
		return
	}
	activityID := uint(atoi(c.Param("activityId")))
	if activityID == 0 {
		response.BadRequest(c, "无效的活动ID")
		return
	}
	page := atoi(c.DefaultQuery("page", "1"))
	limit := atoi(c.DefaultQuery("limit", "20"))
	var statusPtr *int
	if v := c.Query("status"); v != "" {
		val := atoi(v)
		statusPtr = &val
	}

	list, total, err := h.svc.ListActivityRegistrations(storeID, activityID, page, limit, statusPtr)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessWithPagination(c, list, total, page, limit)
}

// RefundActivityRegistration 标记活动报名为已退款（简化版，仅记录，不触发实际退款流水）
// POST /api/v1/stores/:id/activities/:activityId/registrations/:registrationId/refund
func (h *ActivityHandler) RefundActivityRegistration(c *gin.Context) {
	storeID := uint(atoi(c.Param("id")))
	if storeID == 0 {
		response.BadRequest(c, "无效的门店ID")
		return
	}
	activityID := uint(atoi(c.Param("activityId")))
	if activityID == 0 {
		response.BadRequest(c, "无效的活动ID")
		return
	}
	regID := uint(atoi(c.Param("registrationId")))
	if regID == 0 {
		response.BadRequest(c, "无效的报名ID")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var operatorID uint
	if v, ok := c.Get("user_id"); ok {
		if u, ok2 := v.(uint); ok2 {
			operatorID = u
		}
	}
	reg, err := h.svc.RefundActivityRegistration(storeID, activityID, regID, req.Reason, operatorID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, reg)
}

// RegisterActivityWithOrder 用户报名并生成订单（用于需要支付的活动）
// POST /api/v1/activities/:id/register-with-order
func (h *ActivityHandler) RegisterActivityWithOrder(c *gin.Context) {
	activityID := uint(atoi(c.Param("id")))
	if activityID == 0 {
		response.BadRequest(c, "无效的活动ID")
		return
	}
	var req struct {
		Name  string  `json:"name"`
		Phone string  `json:"phone"`
		Fee   float64 `json:"fee"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Name == "" || req.Phone == "" {
		response.BadRequest(c, "姓名和手机号必填")
		return
	}
	if req.Fee < 0 {
		response.BadRequest(c, "报名费用不能为负数")
		return
	}
	var userID uint
	if v, ok := c.Get("user_id"); ok {
		if u, ok2 := v.(uint); ok2 {
			userID = u
		}
	}
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}
	feeDec := decimal.NewFromFloat(req.Fee)
	reg, order, err := h.svc.RegisterActivityWithOrder(activityID, userID, req.Name, req.Phone, feeDec)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"registration": reg, "order": order})
}

// RegisterActivity 用户报名活动（小程序/用户侧）
// POST /api/v1/activities/:id/register
func (h *ActivityHandler) RegisterActivity(c *gin.Context) {
	activityID := uint(atoi(c.Param("id")))
	if activityID == 0 {
		response.BadRequest(c, "无效的活动ID")
		return
	}
	var req struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Name == "" || req.Phone == "" {
		response.BadRequest(c, "姓名和手机号必填")
		return
	}
	var userID uint
	if v, ok := c.Get("user_id"); ok {
		if u, ok2 := v.(uint); ok2 {
			userID = u
		}
	}
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}
	reg, err := h.svc.RegisterActivity(activityID, userID, req.Name, req.Phone)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, reg)
}
