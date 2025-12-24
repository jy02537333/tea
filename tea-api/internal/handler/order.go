package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/internal/service/commission"
	"tea-api/pkg/database"
	"tea-api/pkg/response"
)

// orderService 定义了 OrderHandler 所需的服务方法，便于在测试中注入 fake 实现。
type orderService interface {
	CreateOrderFromCart(userID uint, deliveryType int, addressInfo, remark string, userCouponID uint, storeID uint, orderType int) (*model.Order, error)
	ListOrders(userID uint, status int, page, limit int, storeID uint) ([]model.Order, int64, error)
	GetOrder(userID, orderID uint) (*model.Order, []model.OrderItem, error)
	AdminListOrders(status int, page, limit int, storeID uint, startTime, endTime *time.Time) ([]model.Order, int64, error)
	GetOrderAdmin(orderID uint) (*model.Order, []model.OrderItem, error)
	CancelOrder(userID, orderID uint, reason string) error
	MarkPaid(userID, orderID uint) error
	StartDelivery(userID, orderID uint) error
	Complete(userID, orderID uint) error
	Receive(userID, orderID uint) error
	AdminCancelOrder(orderID uint, reason string) error
	AdminRefundOrder(orderID uint, reason string) error
	AdminRefundStart(orderID uint, reason string) error
	AdminRefundConfirm(orderID uint, reason string) error
	AdminListStoreOrders(storeID uint, status int, page, limit int, startTime, endTime *time.Time, orderID uint) ([]model.Order, int64, error)
}

type OrderHandler struct {
	svc orderService
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

// availableCouponsReq 查询当前订单可用优惠券的请求体（最小版，仅按金额与门店过滤）
type availableCouponsReq struct {
	OrderAmount string `json:"order_amount" binding:"required"` // 本次订单商品总额（未扣券），字符串金额
	StoreID     uint   `json:"store_id"`                        // 可选，门店订单时传入
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

// AvailableCoupons 查询当前订单可用优惠券（最小版：基于用户当前未使用且在有效期内的券，按订单金额与门店做二次过滤）
// 路由建议：POST /api/v1/orders/available-coupons
func (h *OrderHandler) AvailableCoupons(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))

	var req availableCouponsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	db := database.GetDB()
	var userCoupons []model.UserCoupon
	// 仅查当前用户、未使用、券启用且在有效期内的记录
	if err := db.Preload("Coupon").
		Joins("JOIN coupons ON coupons.id = user_coupons.coupon_id").
		Where("user_coupons.user_id = ? AND user_coupons.status = 1 AND coupons.status = 1 AND ? BETWEEN coupons.start_time AND coupons.end_time", userID, time.Now()).
		Order("user_coupons.id desc").
		Find(&userCoupons).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 金额与门店筛选逻辑：
	// - 若 coupon.min_amount > 0，则要求 order_amount >= min_amount；
	// - 平台券(store_id 为空)不过滤门店；门店券仅当 coupon.store_id == req.StoreID 时可用。
	available := make([]model.UserCoupon, 0, len(userCoupons))
	unavailable := make([]gin.H, 0)

	orderAmtDec, err := decimal.NewFromString(req.OrderAmount)
	if err != nil {
		response.BadRequest(c, "order_amount 金额格式不正确")
		return
	}

	for _, uc := range userCoupons {
		coup := uc.Coupon
		// 门店约束
		if coup.StoreID != nil && req.StoreID != 0 && *coup.StoreID != req.StoreID {
			unavailable = append(unavailable, gin.H{"user_coupon_id": uc.ID, "reason": "仅限对应门店订单使用"})
			continue
		}
		// 金额门槛
		if coup.MinAmount.GreaterThan(decimal.Zero) && orderAmtDec.LessThan(coup.MinAmount) {
			unavailable = append(unavailable, gin.H{"user_coupon_id": uc.ID, "reason": "未满足优惠券使用门槛"})
			continue
		}
		available = append(available, uc)
	}

	response.Success(c, gin.H{
		"available":   available,
		"unavailable": unavailable,
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
	var startTimePtr, endTimePtr *time.Time
	if v := c.Query("start_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			startTimePtr = &t
		} else {
			response.BadRequest(c, "start_time 格式错误，应为 RFC3339")
			return
		}
	}
	if v := c.Query("end_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			endTimePtr = &t
		} else {
			response.BadRequest(c, "end_time 格式错误，应为 RFC3339")
			return
		}
	}
	orders, total, err := h.svc.AdminListOrders(status, page, limit, storeID, startTimePtr, endTimePtr)
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
	var startTimePtr, endTimePtr *time.Time
	if v := c.Query("start_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			startTimePtr = &t
		} else {
			response.BadRequest(c, "start_time 格式错误，应为 RFC3339")
			return
		}
	}
	if v := c.Query("end_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			endTimePtr = &t
		} else {
			response.BadRequest(c, "end_time 格式错误，应为 RFC3339")
			return
		}
	}
	// For export, fetch up to 10000 rows
	orders, _, err := h.svc.AdminListOrders(status, 1, 10000, storeID, startTimePtr, endTimePtr)
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

// AdminStoreOrders 管理端按门店维度列出订单列表
// 路由：GET /api/v1/admin/stores/:id/orders
// 支持的查询参数：
// - page: 页码，默认 1
// - page_size 或 limit: 每页大小，默认 20
// - status: 订单状态（可选）
// - start_time, end_time: 时间区间（RFC3339），按创建时间过滤
// - order_id: 订单ID（可选，数字）
func (h *OrderHandler) AdminStoreOrders(c *gin.Context) {
	storeIDVal, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || storeIDVal == 0 {
		response.BadRequest(c, "非法的门店ID")
		return
	}
	storeID := uint(storeIDVal)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSizeStr := c.DefaultQuery("page_size", "")
	if pageSizeStr == "" {
		pageSizeStr = c.DefaultQuery("limit", "20")
	}
	limit, _ := strconv.Atoi(pageSizeStr)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))

	var startTimePtr, endTimePtr *time.Time
	if v := c.Query("start_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			startTimePtr = &t
		} else {
			response.BadRequest(c, "start_time 格式错误，应为 RFC3339")
			return
		}
	}
	if v := c.Query("end_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			endTimePtr = &t
		} else {
			response.BadRequest(c, "end_time 格式错误，应为 RFC3339")
			return
		}
	}

	var orderID uint
	if v := c.Query("order_id"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 32); err == nil {
			orderID = uint(n)
		} else {
			response.BadRequest(c, "order_id 必须为数字")
			return
		}
	}

	orders, total, err := h.svc.AdminListStoreOrders(storeID, status, page, limit, startTimePtr, endTimePtr, orderID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessWithPagination(c, orders, total, page, limit)
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

// AdminAccept 门店/管理员接受订单（占位实现：记录操作日志并返回 ok，不更改订单状态）
func (h *OrderHandler) AdminAccept(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	operatorID := uint(uidVal.(uint))
	oid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || oid == 0 {
		response.BadRequest(c, "非法的订单ID")
		return
	}
	var req struct {
		Note string `json:"note"`
	}
	_ = c.ShouldBindJSON(&req)
	_ = writeOpLog(c, operatorID, "order", "order.admin_accept", map[string]any{
		"order_id": uint(oid),
		"note":     req.Note,
	})
	response.Success(c, gin.H{"ok": true, "status": "accepted"})
}

// AdminReject 门店/管理员拒绝订单（占位实现：记录操作日志并返回 ok，不更改订单状态）
func (h *OrderHandler) AdminReject(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	operatorID := uint(uidVal.(uint))
	oid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || oid == 0 {
		response.BadRequest(c, "非法的订单ID")
		return
	}
	var req struct {
		Reason string `json:"reason"`
		Note   string `json:"note"`
	}
	_ = c.ShouldBindJSON(&req)
	_ = writeOpLog(c, operatorID, "order", "order.admin_reject", map[string]any{
		"order_id": uint(oid),
		"reason":   req.Reason,
		"note":     req.Note,
	})
	response.Success(c, gin.H{"ok": true, "status": "rejected"})
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

	// 退款完成后尝试回滚该订单对应的未提现佣金
	// 若回滚失败，不阻塞退款，仅写一条操作日志供财务后续人工处理
	var opIDPtr *uint
	if operatorID != 0 {
		opIDPtr = &operatorID
	}
	rollbackNote := fmt.Sprintf("order refund confirm: %s", req.Reason)
	if _, err := commission.ReverseOrderCommissions(uint(oid), opIDPtr, rollbackNote); err != nil {
		_ = writeOpLog(c, operatorID, "finance", "commission.rollback_failed", map[string]any{
			"order_id": uint(oid),
			"reason":   req.Reason,
			"error":    err.Error(),
		})
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
