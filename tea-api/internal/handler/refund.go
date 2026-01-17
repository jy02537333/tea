package handler

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

type RefundHandler struct{}

func NewRefundHandler() *RefundHandler { return &RefundHandler{} }

// GET /api/v1/stores/:id/refunds
// 门店侧退款记录列表（需登录+锁店），按门店筛选，支持查询参数：order_id, refund_no(模糊), status, start(创建时间), end(创建时间), page, limit
func (h *RefundHandler) ListStoreRefunds(c *gin.Context) {
	sidStr := strings.TrimSpace(c.Param("id"))
	if sidStr == "" {
		utils.InvalidParam(c, "store id missing")
		return
	}
	sid64, err := strconv.ParseUint(sidStr, 10, 64)
	if err != nil || sid64 == 0 {
		utils.InvalidParam(c, "store id invalid")
		return
	}
	storeID := uint(sid64)

	orderID := strings.TrimSpace(c.Query("order_id"))
	refundNo := strings.TrimSpace(c.Query("refund_no"))
	status := strings.TrimSpace(c.Query("status"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))

	page := toIntRef(c.DefaultQuery("page", "1"))
	size := toIntRef(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}

	db := database.GetDB()
	q := db.Model(&model.Refund{}).
		Joins("JOIN orders ON orders.id = refunds.order_id").
		Where("orders.store_id = ?", storeID).
		Preload("Order").Preload("Payment")

	if orderID != "" {
		if oid, err := strconv.ParseUint(orderID, 10, 64); err == nil {
			q = q.Where("refunds.order_id = ?", uint(oid))
		} else {
			utils.InvalidParam(c, "order_id 参数非法")
			return
		}
	}
	if refundNo != "" {
		q = q.Where("refunds.refund_no LIKE ?", "%"+refundNo+"%")
	}
	if status != "" {
		if st, err := strconv.ParseInt(status, 10, 64); err == nil {
			q = q.Where("refunds.status = ?", int(st))
		} else {
			utils.InvalidParam(c, "status 参数非法")
			return
		}
	}
	if start != "" {
		q = q.Where("refunds.created_at >= ?", start)
	}
	if end != "" {
		q = q.Where("refunds.created_at <= ?", end)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	var list []model.Refund
	if err := q.Order("refunds.id desc").Limit(size).Offset((page - 1) * size).Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.PageSuccess(c, list, total, page, size)
}

// GET /api/v1/admin/refunds
// 支持查询参数：order_id, payment_id, refund_no(模糊), status, start(创建时间), end(创建时间), page, limit
func (h *RefundHandler) ListRefunds(c *gin.Context) {
	orderID := strings.TrimSpace(c.Query("order_id"))
	paymentID := strings.TrimSpace(c.Query("payment_id"))
	refundNo := strings.TrimSpace(c.Query("refund_no"))
	status := strings.TrimSpace(c.Query("status"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))

	page := toIntRef(c.DefaultQuery("page", "1"))
	size := toIntRef(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}

	db := database.GetDB()
	q := db.Model(&model.Refund{}).Preload("Order").Preload("Payment")
	if orderID != "" {
		q = q.Where("order_id = ?", orderID)
	}
	if paymentID != "" {
		q = q.Where("payment_id = ?", paymentID)
	}
	if refundNo != "" {
		q = q.Where("refund_no LIKE ?", "%"+refundNo+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if start != "" {
		q = q.Where("created_at >= ?", start)
	}
	if end != "" {
		q = q.Where("created_at <= ?", end)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	var list []model.Refund
	if err := q.Order("id desc").Limit(size).Offset((page - 1) * size).Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.PageSuccess(c, list, total, page, size)
}

// GET /api/v1/admin/refunds/export
// 导出最近最多5000条，支持查询参数同 ListRefunds
func (h *RefundHandler) ExportRefunds(c *gin.Context) {
	orderID := strings.TrimSpace(c.Query("order_id"))
	paymentID := strings.TrimSpace(c.Query("payment_id"))
	refundNo := strings.TrimSpace(c.Query("refund_no"))
	status := strings.TrimSpace(c.Query("status"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "csv")))

	db := database.GetDB()
	q := db.Model(&model.Refund{}).Preload("Order").Preload("Payment")
	if orderID != "" {
		q = q.Where("order_id = ?", orderID)
	}
	if paymentID != "" {
		q = q.Where("payment_id = ?", paymentID)
	}
	if refundNo != "" {
		q = q.Where("refund_no LIKE ?", "%"+refundNo+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if start != "" {
		q = q.Where("created_at >= ?", start)
	}
	if end != "" {
		q = q.Where("created_at <= ?", end)
	}

	var list []model.Refund
	if err := q.Order("id desc").Limit(5000).Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	filename := "refunds_" + time.Now().Format("20060102150405")
	if format == "xlsx" {
		xf := excelize.NewFile()
		sheet := xf.GetSheetName(0)
		headers := []string{"ID", "Refund No", "Order ID", "Order No", "Payment ID", "Payment No", "Refund Amount", "Refund Reason", "Status", "Refunded At", "Created At"}
		for i, hname := range headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			_ = xf.SetCellValue(sheet, col+"1", hname)
		}
		for idx, it := range list {
			orderNo := ""
			if it.Order.ID != 0 {
				orderNo = it.Order.OrderNo
			}
			payNo := ""
			if it.Payment.ID != 0 {
				payNo = it.Payment.PaymentNo
			}
			refundedAt := ""
			if it.RefundedAt != nil {
				refundedAt = it.RefundedAt.Format("2006-01-02 15:04:05")
			}
			row := []any{
				it.ID,
				it.RefundNo,
				it.OrderID,
				orderNo,
				it.PaymentID,
				payNo,
				it.RefundAmount.String(),
				it.RefundReason,
				it.Status,
				refundedAt,
				it.CreatedAt.Format("2006-01-02 15:04:05"),
			}
			for j, v := range row {
				col, _ := excelize.ColumnNumberToName(j + 1)
				_ = xf.SetCellValue(sheet, col+fmt.Sprintf("%d", idx+2), v)
			}
		}
		var buf bytes.Buffer
		if err := xf.Write(&buf); err != nil {
			utils.Error(c, utils.CodeError, err.Error())
			return
		}
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename="+url.PathEscape(filename+".xlsx"))
		_, _ = c.Writer.Write(buf.Bytes())
		return
	}

	// 默认 CSV
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+url.PathEscape(filename+".csv"))
	_, _ = c.Writer.WriteString("id,refund_no,order_id,order_no,payment_id,payment_no,refund_amount,refund_reason,status,refunded_at,created_at\n")
	for _, it := range list {
		orderNo := ""
		if it.Order.ID != 0 {
			orderNo = it.Order.OrderNo
		}
		payNo := ""
		if it.Payment.ID != 0 {
			payNo = it.Payment.PaymentNo
		}
		refundedAt := ""
		if it.RefundedAt != nil {
			refundedAt = it.RefundedAt.Format("2006-01-02 15:04:05")
		}
		line := fmt.Sprintf("%d,%s,%d,%s,%d,%s,%s,%s,%d,%s,%s\n",
			it.ID,
			csvSafeRef(it.RefundNo),
			it.OrderID,
			csvSafeRef(orderNo),
			it.PaymentID,
			csvSafeRef(payNo),
			csvSafeRef(it.RefundAmount.String()),
			csvSafeRef(it.RefundReason),
			it.Status,
			csvSafeRef(refundedAt),
			it.CreatedAt.Format("2006-01-02 15:04:05"),
		)
		_, _ = c.Writer.WriteString(line)
	}
}

func toIntRef(s string) int { var n int; _, _ = fmt.Sscanf(s, "%d", &n); return n }
func csvSafeRef(s string) string {
	s = strings.ReplaceAll(s, ",", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// ===== 用户侧：退款列表查询 =====

// GET /api/v1/refunds
// 用户侧退款列表（需登录），按当前用户筛选，支持 order_id/status/start/end/page/limit
func (h *RefundHandler) ListMyRefunds(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok || uid == 0 {
		utils.Unauthorized(c, "未登录或令牌无效")
		return
	}

	orderID := strings.TrimSpace(c.Query("order_id"))
	status := strings.TrimSpace(c.Query("status"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))

	page := toIntRef(c.DefaultQuery("page", "1"))
	size := toIntRef(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}

	db := database.GetDB()
	q := db.Model(&model.Refund{}).
		Joins("JOIN orders ON orders.id = refunds.order_id").
		Where("orders.user_id = ?", uid).
		Preload("Order").Preload("Payment")

	if orderID != "" {
		// 尝试解析为数字，避免SQL注入与类型不匹配
		if oid, err := strconv.ParseUint(orderID, 10, 64); err == nil {
			q = q.Where("refunds.order_id = ?", uint(oid))
		} else {
			utils.InvalidParam(c, "order_id 参数非法")
			return
		}
	}
	if status != "" {
		if st, err := strconv.ParseInt(status, 10, 64); err == nil {
			q = q.Where("refunds.status = ?", int(st))
		} else {
			utils.InvalidParam(c, "status 参数非法")
			return
		}
	}
	if start != "" {
		q = q.Where("refunds.created_at >= ?", start)
	}
	if end != "" {
		q = q.Where("refunds.created_at <= ?", end)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	var list []model.Refund
	if err := q.Order("refunds.id desc").Limit(size).Offset((page-1)*size).Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.PageSuccess(c, list, total, page, size)
}

// 使用同包内已有的 currentUserID(c)（定义于 withdrawal.go）
