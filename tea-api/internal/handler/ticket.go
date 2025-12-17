package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/response"
)

type TicketHandler struct{ svc *service.TicketService }

func NewTicketHandler() *TicketHandler { return &TicketHandler{svc: service.NewTicketService()} }

// List 管理端工单列表
// GET /api/v1/admin/tickets
func (h *TicketHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := strings.TrimSpace(c.Query("status"))
	ticketType := strings.TrimSpace(c.Query("type"))
	source := strings.TrimSpace(c.Query("source"))
	priority := strings.TrimSpace(c.Query("priority"))
	keyword := strings.TrimSpace(c.Query("keyword"))

	var storeIDPtr, userIDPtr *uint
	if v := strings.TrimSpace(c.Query("store_id")); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil && n > 0 {
			val := uint(n)
			storeIDPtr = &val
		}
	}
	if v := strings.TrimSpace(c.Query("user_id")); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil && n > 0 {
			val := uint(n)
			userIDPtr = &val
		}
	}

	list, total, err := h.svc.ListTickets(page, limit, status, ticketType, source, priority, keyword, storeIDPtr, userIDPtr)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessWithPagination(c, list, total, page, limit)
}

// Get 管理端工单详情
// GET /api/v1/admin/tickets/:id
func (h *TicketHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "非法ID")
		return
	}
	t, err := h.svc.GetTicket(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, t)
}

// Create 管理端创建工单（人工录入）
// POST /api/v1/admin/tickets
func (h *TicketHandler) Create(c *gin.Context) {
	var req struct {
		Type        string `json:"type" binding:"required"`
		Source      string `json:"source" binding:"required"`
		UserID      *uint  `json:"user_id"`
		OrderID     *uint  `json:"order_id"`
		StoreID     *uint  `json:"store_id"`
		Title       string `json:"title" binding:"required"`
		Content     string `json:"content"`
		Attachments string `json:"attachments"`
		Priority    string `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	var operatorID uint
	if v, ok := c.Get("user_id"); ok {
		if u, ok2 := v.(uint); ok2 {
			operatorID = u
		}
	}

	t := &model.Ticket{
		Type:        strings.TrimSpace(req.Type),
		Source:      strings.TrimSpace(req.Source),
		UserID:      req.UserID,
		OrderID:     req.OrderID,
		StoreID:     req.StoreID,
		Title:       strings.TrimSpace(req.Title),
		Content:     strings.TrimSpace(req.Content),
		Attachments: strings.TrimSpace(req.Attachments),
		Priority:    strings.TrimSpace(req.Priority),
	}
	if t.Priority == "" {
		t.Priority = "normal"
	}
	if operatorID != 0 {
		t.CreatedBy = operatorID
		t.UpdatedBy = operatorID
	}
	// CreatedAt/UpdatedAt 由 GORM 默认填充
	if err := h.svc.CreateTicket(t); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, t)
}

// Update 管理端更新工单
// PUT /api/v1/admin/tickets/:id
func (h *TicketHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "非法ID")
		return
	}
	var req struct {
		Status       *string `json:"status"`
		Priority     *string `json:"priority"`
		Remark       *string `json:"remark"`
		RejectReason *string `json:"reject_reason"`
		AssigneeID   *uint   `json:"assignee_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	var operatorID uint
	if v, ok := c.Get("user_id"); ok {
		if u, ok2 := v.(uint); ok2 {
			operatorID = u
		}
	}

	input := service.UpdateTicketInput{
		Status:       req.Status,
		Priority:     req.Priority,
		Remark:       req.Remark,
		RejectReason: req.RejectReason,
		AssigneeID:   req.AssigneeID,
	}
	if err := h.svc.UpdateTicket(uint(id), input, operatorID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 返回更新后的最新数据
	t, err := h.svc.GetTicket(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, t)
}
