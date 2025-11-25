package handler

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

type PaymentAdminHandler struct{}

func NewPaymentAdminHandler() *PaymentAdminHandler { return &PaymentAdminHandler{} }

// GET /api/v1/admin/payments
// 支持查询参数：order_id, store_id, payment_no(模糊), status, method, start(创建时间), end(创建时间), page, limit
func (h *PaymentAdminHandler) ListPayments(c *gin.Context) {
	orderID := strings.TrimSpace(c.Query("order_id"))
	storeID := strings.TrimSpace(c.Query("store_id"))
	paymentNo := strings.TrimSpace(c.Query("payment_no"))
	status := strings.TrimSpace(c.Query("status"))
	method := strings.TrimSpace(c.Query("method"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))

	page := toIntPay(c.DefaultQuery("page", "1"))
	size := toIntPay(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}

	db := database.GetDB()
	q := db.Model(&model.Payment{}).Preload("Order")
	if orderID != "" {
		q = q.Where("order_id = ?", orderID)
	}
	if storeID != "" {
		q = q.Joins("JOIN orders o ON o.id = payments.order_id").Where("o.store_id = ?", storeID)
	}
	if paymentNo != "" {
		q = q.Where("payment_no LIKE ?", "%"+paymentNo+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if method != "" {
		q = q.Where("payment_method = ?", method)
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
	var list []model.Payment
	if err := q.Order("id desc").Limit(size).Offset((page - 1) * size).Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	// 附带门店名称，便于前端展示
	storeIDs := make(map[uint]struct{})
	for _, it := range list {
		if it.Order.ID != 0 {
			storeIDs[it.Order.StoreID] = struct{}{}
		}
	}
	sids := make([]uint, 0, len(storeIDs))
	for id := range storeIDs {
		sids = append(sids, id)
	}
	sname := map[uint]string{}
	if len(sids) > 0 {
		var stores []model.Store
		if err := database.GetDB().Model(&model.Store{}).Where("id IN ?", sids).Find(&stores).Error; err == nil {
			for _, s := range stores {
				sname[s.ID] = s.Name
			}
		}
	}
	type PaymentResp struct {
		model.Payment
		StoreName string `json:"store_name"`
	}
	rows := make([]PaymentResp, 0, len(list))
	for _, it := range list {
		rows = append(rows, PaymentResp{Payment: it, StoreName: sname[it.Order.StoreID]})
	}
	utils.PageSuccess(c, rows, total, page, size)
}

// GET /api/v1/admin/payments/export?format=csv|xlsx
// 导出最近最多5000条，支持查询参数同 ListPayments
func (h *PaymentAdminHandler) ExportPayments(c *gin.Context) {
	orderID := strings.TrimSpace(c.Query("order_id"))
	storeID := strings.TrimSpace(c.Query("store_id"))
	paymentNo := strings.TrimSpace(c.Query("payment_no"))
	status := strings.TrimSpace(c.Query("status"))
	method := strings.TrimSpace(c.Query("method"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "csv")))

	db := database.GetDB()
	q := db.Model(&model.Payment{}).Preload("Order")
	if orderID != "" {
		q = q.Where("order_id = ?", orderID)
	}
	if storeID != "" {
		q = q.Joins("JOIN orders o ON o.id = payments.order_id").Where("o.store_id = ?", storeID)
	}
	if paymentNo != "" {
		q = q.Where("payment_no LIKE ?", "%"+paymentNo+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if method != "" {
		q = q.Where("payment_method = ?", method)
	}
	if start != "" {
		q = q.Where("created_at >= ?", start)
	}
	if end != "" {
		q = q.Where("created_at <= ?", end)
	}

	var list []model.Payment
	if err := q.Order("id desc").Limit(5000).Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	// 查询门店名称映射
	storeIDs := make(map[uint]struct{})
	for _, it := range list {
		if it.Order.ID != 0 {
			storeIDs[it.Order.StoreID] = struct{}{}
		}
	}
	sids := make([]uint, 0, len(storeIDs))
	for id := range storeIDs {
		sids = append(sids, id)
	}
	sname := map[uint]string{}
	if len(sids) > 0 {
		var stores []model.Store
		if err := database.GetDB().Model(&model.Store{}).Where("id IN ?", sids).Find(&stores).Error; err == nil {
			for _, s := range stores {
				sname[s.ID] = s.Name
			}
		}
	}

	filename := "payments_" + time.Now().Format("20060102150405")
	if format == "xlsx" {
		xf := excelize.NewFile()
		sheet := xf.GetSheetName(0)
		headers := []string{"ID", "Payment No", "Order ID", "Order No", "Store ID", "Store Name", "Method", "Amount", "Status", "Paid At", "Created At"}
		for i, hname := range headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			_ = xf.SetCellValue(sheet, col+"1", hname)
		}
		for idx, it := range list {
			orderNo := ""
			if it.Order.ID != 0 {
				orderNo = it.Order.OrderNo
			}
			paidAt := ""
			if it.PaidAt != nil {
				paidAt = it.PaidAt.Format("2006-01-02 15:04:05")
			}
			row := []any{
				it.ID,
				it.PaymentNo,
				it.OrderID,
				orderNo,
				it.Order.StoreID,
				sname[it.Order.StoreID],
				it.PaymentMethod,
				it.Amount.String(),
				it.Status,
				paidAt,
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
	_, _ = c.Writer.WriteString("id,payment_no,order_id,order_no,store_id,store_name,method,amount,status,paid_at,created_at\n")
	for _, it := range list {
		orderNo := ""
		if it.Order.ID != 0 {
			orderNo = it.Order.OrderNo
		}
		paidAt := ""
		if it.PaidAt != nil {
			paidAt = it.PaidAt.Format("2006-01-02 15:04:05")
		}
		line := fmt.Sprintf("%d,%s,%d,%s,%d,%s,%d,%s,%d,%s,%s\n",
			it.ID,
			csvSafePay(it.PaymentNo),
			it.OrderID,
			csvSafePay(orderNo),
			it.Order.StoreID,
			csvSafePay(sname[it.Order.StoreID]),
			it.PaymentMethod,
			csvSafePay(it.Amount.String()),
			it.Status,
			csvSafePay(paidAt),
			it.CreatedAt.Format("2006-01-02 15:04:05"),
		)
		_, _ = c.Writer.WriteString(line)
	}
}

func toIntPay(s string) int { var n int; _, _ = fmt.Sscanf(s, "%d", &n); return n }
func csvSafePay(s string) string {
	s = strings.ReplaceAll(s, ",", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
