package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

// DashboardHandler 提供后台首页相关接口
// 包含待办统计等（例如待处理工单数）。

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler { return &DashboardHandler{} }

// Todos 返回后台首页待办统计信息
// GET /api/v1/admin/dashboard/todos
// 当前实现：统计待处理工单、待发货订单、待处理提现等数量。
func (h *DashboardHandler) Todos(c *gin.Context) {
	db := database.GetDB()

	var ticketPending int64
	_ = db.Model(&model.Ticket{}).
		Where("status IN (?)", []string{"new", "pending", "waiting_customer"}).
		Count(&ticketPending).Error

	var orderToShip int64
	_ = db.Model(&model.Order{}).
		Where("status = ? AND pay_status = ?", 2, 2).
		Count(&orderToShip).Error

	var withdrawPending int64
	_ = db.Model(&model.WithdrawRecord{}).
		Where("status IN (?)", []int{model.WithdrawStatusPending, model.WithdrawStatusProcessing}).
		Count(&withdrawPending).Error

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"ticket_pending_count":   ticketPending,
			"order_to_ship_count":    orderToShip,
			"withdraw_pending_count": withdrawPending,
		},
	})
}
