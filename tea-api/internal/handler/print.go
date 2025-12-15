package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

// PrintHandler 打印任务处理器
type PrintHandler struct{}

// NewPrintHandler 创建打印处理器
func NewPrintHandler() *PrintHandler {
	return &PrintHandler{}
}

// PrintJobRequest 打印任务请求
type PrintJobRequest struct {
	StoreID      uint   `json:"store_id" binding:"required"`
	OrderID      uint   `json:"order_id" binding:"required"`
	TemplateType string `json:"template_type" binding:"required"` // kitchen|receipt|label
	Copies       int    `json:"copies"`
}

// CreatePrintJob 创建打印任务
// POST /api/v1/stores/:store_id/print/jobs
func (h *PrintHandler) CreatePrintJob(c *gin.Context) {
	storeIDStr := c.Param("store_id")
	storeID, err := strconv.ParseUint(storeIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "invalid store_id"})
		return
	}

	var req PrintJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	// 确保 storeID 匹配
	if uint64(req.StoreID) != storeID {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "store_id mismatch"})
		return
	}

	db := database.GetDB()

	// 查询订单信息
	var order model.Order
	if err := db.First(&order, req.OrderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 4004, "message": "order not found"})
		return
	}

	// 验证订单是否属于该门店
	if order.StoreID == nil || *order.StoreID != req.StoreID {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4003, "message": "order does not belong to this store"})
		return
	}

	// 设置默认打印份数
	if req.Copies <= 0 {
		req.Copies = 1
	}

	// 创建打印任务记录（简化实现，实际应该有专门的打印任务表）
	// 这里我们使用一个简单的方式：通过 RabbitMQ 或 Redis 队列发送打印任务
	// 暂时返回成功响应，实际打印逻辑由打印服务处理

	printJob := gin.H{
		"job_id":        generatePrintJobID(),
		"store_id":      req.StoreID,
		"order_id":      req.OrderID,
		"order_no":      order.OrderNo,
		"template_type": req.TemplateType,
		"copies":        req.Copies,
		"status":        "pending",
		"created_at":    time.Now().Format(time.RFC3339),
	}

	// TODO: 实际实现中应该将打印任务推送到队列
	// 例如：publishToPrintQueue(printJob)

	c.JSON(http.StatusCreated, gin.H{
		"code": 0,
		"message": "print job created",
		"data": printJob,
	})
}

// AcceptOrder 门店接单
// POST /api/v1/stores/:store_id/orders/:order_id/accept
func (h *PrintHandler) AcceptOrder(c *gin.Context) {
	storeIDStr := c.Param("store_id")
	orderIDStr := c.Param("order_id")
	
	storeID, err := strconv.ParseUint(storeIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "invalid store_id"})
		return
	}

	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "invalid order_id"})
		return
	}

	db := database.GetDB()

	// 查询订单
	var order model.Order
	if err := db.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 4004, "message": "order not found"})
		return
	}

	// 验证订单是否属于该门店
	if order.StoreID == nil || *order.StoreID != uint(storeID) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4003, "message": "order does not belong to this store"})
		return
	}

	// 验证订单状态（只有已支付的订单才能接单）
	if order.Status != "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "order status is not paid"})
		return
	}

	// 更新订单状态为已接单
	if err := db.Model(&order).Update("status", "store_accepted").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	// 记录订单日志
	orderLog := model.OrderLog{
		OrderID:     uint(orderID),
		Status:      "store_accepted",
		Description: "门店已接单",
	}
	db.Create(&orderLog)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "order accepted",
		"data": gin.H{
			"order_id": orderID,
			"status":   "store_accepted",
		},
	})
}

// RejectOrder 门店拒单
// POST /api/v1/stores/:store_id/orders/:order_id/reject
func (h *PrintHandler) RejectOrder(c *gin.Context) {
	storeIDStr := c.Param("store_id")
	orderIDStr := c.Param("order_id")
	
	storeID, err := strconv.ParseUint(storeIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "invalid store_id"})
		return
	}

	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "invalid order_id"})
		return
	}

	type RejectRequest struct {
		Reason string `json:"reason" binding:"required"`
	}

	var req RejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	db := database.GetDB()

	// 查询订单
	var order model.Order
	if err := db.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 4004, "message": "order not found"})
		return
	}

	// 验证订单是否属于该门店
	if order.StoreID == nil || *order.StoreID != uint(storeID) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4003, "message": "order does not belong to this store"})
		return
	}

	// 验证订单状态
	if order.Status != "paid" && order.Status != "pending_payment" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "order cannot be rejected in current status"})
		return
	}

	// 更新订单状态为已拒绝
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status": "rejected",
		"note":   req.Reason,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	// 记录订单日志
	orderLog := model.OrderLog{
		OrderID:     uint(orderID),
		Status:      "rejected",
		Description: "门店拒单: " + req.Reason,
	}
	tx.Create(&orderLog)

	// 如果订单已支付，需要退款
	if order.Status == "paid" && order.PaidAmount.GreaterThan(model.DecimalZero) {
		// TODO: 触发退款流程
		// 这里应该调用退款服务
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "order rejected",
		"data": gin.H{
			"order_id": orderID,
			"status":   "rejected",
			"reason":   req.Reason,
		},
	})
}

// ListStoreOrders 获取门店订单列表
// GET /api/v1/stores/:store_id/orders
func (h *PrintHandler) ListStoreOrders(c *gin.Context) {
	storeIDStr := c.Param("store_id")
	storeID, err := strconv.ParseUint(storeIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "invalid store_id"})
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	status := c.Query("status")

	db := database.GetDB()
	query := db.Model(&model.Order{}).Where("store_id = ?", storeID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	var total int64
	query.Count(&total)

	// 查询列表
	var orders []model.Order
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"items": orders,
			"total": total,
			"page":  page,
			"size":  pageSize,
		},
	})
}

// generatePrintJobID 生成打印任务ID
func generatePrintJobID() string {
	return "PJ" + time.Now().Format("20060102150405") + strconv.FormatInt(time.Now().UnixNano()%10000, 10)
}
