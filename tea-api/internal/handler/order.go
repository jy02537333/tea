package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/database"
	"tea-api/pkg/response"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler() *OrderHandler { return &OrderHandler{svc: service.NewOrderService()} }

type createOrderReq struct {
	DeliveryType int    `json:"delivery_type" binding:"required"` // 1自取 2配送
	AddressInfo  string `json:"address_info"`
	Remark       string `json:"remark"`
	UserCouponID uint   `json:"user_coupon_id"`
	StoreID      uint   `json:"store_id"`
	OrderType    int    `json:"order_type"` // 1商城 2堂食 3外卖
}

// CreateFromCart 从购物车下单
func (h *OrderHandler) CreateFromCart(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))

	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	order, err := h.svc.CreateOrderFromCart(userID, req.DeliveryType, req.AddressInfo, req.Remark, req.UserCouponID, req.StoreID, req.OrderType)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	// 返回精简字段并将金额转为数字，便于前端与测试解析
	payAmt, _ := order.PayAmount.Float64()
	discAmt, _ := order.DiscountAmount.Float64()
	response.Success(c, gin.H{
		"id":              order.ID,
		"order_no":        order.OrderNo,
		"pay_amount":      payAmt,
		"discount_amount": discAmt,
	})
}

// List 列表
func (h *OrderHandler) List(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	var storeID uint
	if v := c.Query("store_id"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 32); err == nil {
			storeID = uint(n)
		}
	}

	orders, total, err := h.svc.ListOrders(userID, status, page, limit, storeID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessWithPagination(c, orders, total, page, limit)
}

// Detail 详情
func (h *OrderHandler) Detail(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))
	oid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法的订单ID")
		return
	}
	order, items, err := h.svc.GetOrder(userID, uint(oid))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"order": order, "items": items})
}

// AdminList 管理端列出订单（需 admin 权限）
func (h *OrderHandler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	var storeID uint
	if v := c.Query("store_id"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 32); err == nil {
			storeID = uint(n)
		}
	}
	orders, total, err := h.svc.AdminListOrders(status, page, limit, storeID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessWithPagination(c, orders, total, page, limit)
}

// AdminExport 管理端导出订单（CSV）
func (h *OrderHandler) AdminExport(c *gin.Context) {
	// reuse AdminListOrders to fetch all (limit can be large or pagination removed for export)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	var storeID uint
	if v := c.Query("store_id"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 32); err == nil {
			storeID = uint(n)
		}
	}
	// For export, fetch up to 10000 rows
	orders, _, err := h.svc.AdminListOrders(status, 1, 10000, storeID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	// build CSV
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=orders_export.csv")
	w := c.Writer
	// write header
	_, _ = w.Write([]byte("id,order_no,store_id,user_id,pay_amount,status,pay_status,created_at\n"))
	for _, o := range orders {
		line := fmt.Sprintf("%d,%s,%d,%d,%.2f,%d,%d,%s\n", o.ID, o.OrderNo, o.StoreID, o.UserID, func() float64 { f, _ := o.PayAmount.Float64(); return f }(), o.Status, o.PayStatus, o.CreatedAt.Format("2006-01-02 15:04:05"))
		_, _ = w.Write([]byte(line))
	}
}

// AdminDetail 管理端获取任意订单详情（需 admin 权限）
func (h *OrderHandler) AdminDetail(c *gin.Context) {
	oid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法的订单ID")
		return
	}
	order, items, err := h.svc.GetOrderAdmin(uint(oid))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"order": order, "items": items})
}

// Cancel 取消
func (h *OrderHandler) Cancel(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))
	oid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法的订单ID")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.svc.CancelOrder(userID, uint(oid), req.Reason); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// Pay 模拟支付
func (h *OrderHandler) Pay(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))
	oid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法的订单ID")
		return
	}
	if err := h.svc.MarkPaid(userID, uint(oid)); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// Deliver 发货
func (h *OrderHandler) Deliver(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))
	oid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法的订单ID")
		return
	}
	if err := h.svc.StartDelivery(userID, uint(oid)); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// Complete 完成
func (h *OrderHandler) Complete(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))
	oid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法的订单ID")
		return
	}
	if err := h.svc.Complete(userID, uint(oid)); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// Receive 用户确认收货/完成订单（仅限本人）
func (h *OrderHandler) Receive(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))
	oid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法的订单ID")
		return
	}
	if err := h.svc.Receive(userID, uint(oid)); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// AdminCancel 管理员取消订单（需权限）
func (h *OrderHandler) AdminCancel(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	operatorID := uint(uidVal.(uint))
	oid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法的订单ID")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	// 记录状态前后对比
	var before model.Order
	_ = database.GetDB().First(&before, uint(oid)).Error
	if err := h.svc.AdminCancelOrder(uint(oid), req.Reason); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	var after model.Order
	_ = database.GetDB().First(&after, uint(oid)).Error
	// 写入操作日志（详细）
	_ = writeOpLog(c, operatorID, "order", "order.admin_cancel", map[string]any{
		"order_id": uint(oid),
		"reason":   req.Reason,
		"before":   map[string]any{"status": before.Status, "pay_status": before.PayStatus},
		"after":    map[string]any{"status": after.Status, "pay_status": after.PayStatus},
	})
	response.Success(c, gin.H{"ok": true})
}

// AdminRefund 管理端手动退款（需权限）
func (h *OrderHandler) AdminRefund(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	operatorID := uint(uidVal.(uint))
	oid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法的订单ID")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	var before model.Order
	_ = database.GetDB().First(&before, uint(oid)).Error
	if err := h.svc.AdminRefundOrder(uint(oid), req.Reason); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	var after model.Order
	_ = database.GetDB().First(&after, uint(oid)).Error
	_ = writeOpLog(c, operatorID, "finance", "order.refund", map[string]any{
		"order_id": uint(oid),
		"reason":   req.Reason,
		"before":   map[string]any{"status": before.Status, "pay_status": before.PayStatus},
		"after":    map[string]any{"status": after.Status, "pay_status": after.PayStatus},
	})
	response.Success(c, gin.H{"ok": true})
}

// AdminRefundStart 标记退款中（需权限）
func (h *OrderHandler) AdminRefundStart(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	operatorID := uint(uidVal.(uint))
	oid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法的订单ID")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	var before model.Order
	_ = database.GetDB().First(&before, uint(oid)).Error
	if err := h.svc.AdminRefundStart(uint(oid), req.Reason); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	var after model.Order
	_ = database.GetDB().First(&after, uint(oid)).Error
	_ = writeOpLog(c, operatorID, "finance", "order.refund_start", map[string]any{
		"order_id": uint(oid),
		"reason":   req.Reason,
		"before":   map[string]any{"status": before.Status, "pay_status": before.PayStatus},
		"after":    map[string]any{"status": after.Status, "pay_status": after.PayStatus},
	})
	response.Success(c, gin.H{"ok": true})
}

// AdminRefundConfirm 确认退款完成（需权限）
func (h *OrderHandler) AdminRefundConfirm(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	operatorID := uint(uidVal.(uint))
	oid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法的订单ID")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	var before model.Order
	_ = database.GetDB().First(&before, uint(oid)).Error
	if err := h.svc.AdminRefundConfirm(uint(oid), req.Reason); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	var after model.Order
	_ = database.GetDB().First(&after, uint(oid)).Error
	_ = writeOpLog(c, operatorID, "finance", "order.refund_confirm", map[string]any{
		"order_id": uint(oid),
		"reason":   req.Reason,
		"before":   map[string]any{"status": before.Status, "pay_status": before.PayStatus},
		"after":    map[string]any{"status": after.Status, "pay_status": after.PayStatus},
	})
	response.Success(c, gin.H{"ok": true})
}

// writeOpLog 写入操作日志（包含详细 requestData）
func writeOpLog(c *gin.Context, userID uint, module, operation string, data map[string]any) error {
	bs, _ := json.Marshal(data)
	rec := &model.OperationLog{
		UserID:      userID,
		Module:      module,
		Operation:   operation,
		Description: "",
		RequestData: string(bs),
		IP:          c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
	}
	return database.GetDB().Create(rec).Error
}
