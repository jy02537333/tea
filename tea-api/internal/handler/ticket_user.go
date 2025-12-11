package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tea-api/internal/model"
	"tea-api/internal/service"
)

// TicketUserHandler 提供用户侧（小程序）工单接口
// 仅支持创建工单，供意见反馈/订单投诉等入口使用。
type TicketUserHandler struct {
	svc *service.TicketService
}

func NewTicketUserHandler() *TicketUserHandler {
	return &TicketUserHandler{svc: service.NewTicketService()}
}

// Create 允许登录用户创建工单
// POST /api/v1/tickets
func (h *TicketUserHandler) Create(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid user"})
		return
	}

	var req struct {
		Type        string `json:"type" binding:"required"`
		Source      string `json:"source" binding:"required"`
		OrderID     *uint  `json:"order_id"`
		StoreID     *uint  `json:"store_id"`
		Title       string `json:"title" binding:"required"`
		Content     string `json:"content" binding:"required"`
		Attachments string `json:"attachments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
		return
	}

	ticket := &model.Ticket{
		Type:        req.Type,
		Source:      req.Source,
		UserID:      &userID,
		OrderID:     req.OrderID,
		StoreID:     req.StoreID,
		Title:       req.Title,
		Content:     req.Content,
		Attachments: req.Attachments,
		Status:      "new",
		Priority:    "normal",
	}

	if err := h.svc.CreateTicket(ticket); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ticket})
}
