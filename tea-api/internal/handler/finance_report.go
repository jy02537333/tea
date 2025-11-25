package handler

import (
	"bytes"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"

	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

type FinanceReportHandler struct{}

func NewFinanceReportHandler() *FinanceReportHandler { return &FinanceReportHandler{} }

type financeSummary struct {
	TotalPaymentsCount int64           `json:"total_payments_count"`
	TotalPaymentsAmt   decimal.Decimal `json:"total_payments_amount"`
	TotalRefundsCount  int64           `json:"total_refunds_count"`
	TotalRefundsAmt    decimal.Decimal `json:"total_refunds_amount"`
	NetAmount          decimal.Decimal `json:"net_amount"`
}

type dailyRow struct {
	Date        string          `json:"date"`
	PayCount    int64           `json:"pay_count"`
	PayAmount   decimal.Decimal `json:"pay_amount"`
	RefundCount int64           `json:"refund_count"`
	RefundAmt   decimal.Decimal `json:"refund_amount"`
	NetAmount   decimal.Decimal `json:"net_amount"`
}

// GET /api/v1/admin/finance/summary
// 查询参数：start(含) end(含)、group=day|store|method 可选；过滤：store_id、method
func (h *FinanceReportHandler) Summary(c *gin.Context) {
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))
	group := strings.TrimSpace(c.DefaultQuery("group", ""))
	storeID := strings.TrimSpace(c.Query("store_id"))
	method := strings.TrimSpace(c.Query("method"))

	db := database.GetDB()

	var pays []model.Payment
	pq := db.Model(&model.Payment{}).Preload("Order").Where("status = ?", 2)
	if start != "" {
		pq = pq.Where("created_at >= ?", start)
	}
	if end != "" {
		pq = pq.Where("created_at <= ?", end)
	}
	if storeID != "" {
		pq = pq.Joins("JOIN orders o ON o.id = payments.order_id").Where("o.store_id = ?", storeID)
	}
	if method != "" {
		pq = pq.Where("payment_method = ?", method)
	}
	if err := pq.Find(&pays).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	var refs []model.Refund
	rq := db.Model(&model.Refund{}).Preload("Order").Preload("Payment").Where("status = ?", 2)
	if start != "" {
		rq = rq.Where("created_at >= ?", start)
	}
	if end != "" {
		rq = rq.Where("created_at <= ?", end)
	}
	if storeID != "" {
		rq = rq.Joins("JOIN orders o ON o.id = refunds.order_id").Where("o.store_id = ?", storeID)
	}
	if method != "" {
		rq = rq.Joins("JOIN payments p ON p.id = refunds.payment_id").Where("p.payment_method = ?", method)
	}
	if err := rq.Find(&refs).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	sum := financeSummary{TotalPaymentsAmt: decimal.NewFromInt(0), TotalRefundsAmt: decimal.NewFromInt(0)}
	for _, p := range pays {
		sum.TotalPaymentsCount++
		sum.TotalPaymentsAmt = sum.TotalPaymentsAmt.Add(p.Amount)
	}
	for _, r := range refs {
		sum.TotalRefundsCount++
		sum.TotalRefundsAmt = sum.TotalRefundsAmt.Add(r.RefundAmount)
	}
	sum.NetAmount = sum.TotalPaymentsAmt.Sub(sum.TotalRefundsAmt)

	if group == "" {
		utils.Success(c, gin.H{"summary": sum})
		return
	}

	if group == "day" {
		// 按日分组
		dayMap := map[string]*dailyRow{}
		dateKey := func(t time.Time) string { return t.Format("2006-01-02") }
		for _, p := range pays {
			d := dateKey(p.CreatedAt)
			row := dayMap[d]
			if row == nil {
				row = &dailyRow{Date: d, PayAmount: decimal.NewFromInt(0), RefundAmt: decimal.NewFromInt(0), NetAmount: decimal.NewFromInt(0)}
				dayMap[d] = row
			}
			row.PayCount++
			row.PayAmount = row.PayAmount.Add(p.Amount)
		}
		for _, r := range refs {
			d := dateKey(r.CreatedAt)
			row := dayMap[d]
			if row == nil {
				row = &dailyRow{Date: d, PayAmount: decimal.NewFromInt(0), RefundAmt: decimal.NewFromInt(0), NetAmount: decimal.NewFromInt(0)}
				dayMap[d] = row
			}
			row.RefundCount++
			row.RefundAmt = row.RefundAmt.Add(r.RefundAmount)
		}
		days := make([]string, 0, len(dayMap))
		for d := range dayMap {
			days = append(days, d)
		}
		sort.Strings(days)
		rows := make([]dailyRow, 0, len(days))
		for _, d := range days {
			r := dayMap[d]
			r.NetAmount = r.PayAmount.Sub(r.RefundAmt)
			rows = append(rows, *r)
		}
		utils.Success(c, gin.H{"summary": sum, "rows": rows})
		return
	}

	if group == "store" {
		type storeRow struct {
			StoreID                         uint
			PayCount                        int64
			PayAmount, RefundAmt, NetAmount decimal.Decimal
			RefundCount                     int64
		}
		m := map[uint]*storeRow{}
		for _, p := range pays {
			sid := p.Order.StoreID
			row := m[sid]
			if row == nil {
				row = &storeRow{StoreID: sid, PayAmount: decimal.NewFromInt(0), RefundAmt: decimal.NewFromInt(0)}
				m[sid] = row
			}
			row.PayCount++
			row.PayAmount = row.PayAmount.Add(p.Amount)
		}
		for _, r := range refs {
			sid := r.Order.StoreID
			row := m[sid]
			if row == nil {
				row = &storeRow{StoreID: sid, PayAmount: decimal.NewFromInt(0), RefundAmt: decimal.NewFromInt(0)}
				m[sid] = row
			}
			row.RefundCount++
			row.RefundAmt = row.RefundAmt.Add(r.RefundAmount)
		}
		keys := make([]uint, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		// 获取门店名称映射
		var stores []model.Store
		if len(keys) > 0 {
			if err := db.Model(&model.Store{}).Where("id IN ?", keys).Find(&stores).Error; err != nil {
				utils.Error(c, utils.CodeError, err.Error())
				return
			}
		}
		sname := map[uint]string{}
		for _, s := range stores {
			sname[s.ID] = s.Name
		}
		var rows []gin.H
		for _, k := range keys {
			row := m[k]
			row.NetAmount = row.PayAmount.Sub(row.RefundAmt)
			rows = append(rows, gin.H{"store_id": k, "store_name": sname[k], "pay_count": row.PayCount, "pay_amount": row.PayAmount, "refund_count": row.RefundCount, "refund_amount": row.RefundAmt, "net_amount": row.NetAmount})
		}
		utils.Success(c, gin.H{"summary": sum, "rows": rows})
		return
	}

	if group == "method" {
		type methodRow struct {
			Method                          int
			PayCount                        int64
			PayAmount, RefundAmt, NetAmount decimal.Decimal
			RefundCount                     int64
		}
		m := map[int]*methodRow{}
		for _, p := range pays {
			md := p.PaymentMethod
			row := m[md]
			if row == nil {
				row = &methodRow{Method: md, PayAmount: decimal.NewFromInt(0), RefundAmt: decimal.NewFromInt(0)}
				m[md] = row
			}
			row.PayCount++
			row.PayAmount = row.PayAmount.Add(p.Amount)
		}
		for _, r := range refs {
			md := 0
			if r.Payment.ID != 0 {
				md = r.Payment.PaymentMethod
			}
			row := m[md]
			if row == nil {
				row = &methodRow{Method: md, PayAmount: decimal.NewFromInt(0), RefundAmt: decimal.NewFromInt(0)}
				m[md] = row
			}
			row.RefundCount++
			row.RefundAmt = row.RefundAmt.Add(r.RefundAmount)
		}
		keys := make([]int, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		var rows []gin.H
		for _, k := range keys {
			row := m[k]
			row.NetAmount = row.PayAmount.Sub(row.RefundAmt)
			rows = append(rows, gin.H{"method": k, "pay_count": row.PayCount, "pay_amount": row.PayAmount, "refund_count": row.RefundCount, "refund_amount": row.RefundAmt, "net_amount": row.NetAmount})
		}
		utils.Success(c, gin.H{"summary": sum, "rows": rows})
		return
	}

	utils.Success(c, gin.H{"summary": sum})
}

// GET /api/v1/admin/finance/summary/export?format=csv|xlsx&group=day
func (h *FinanceReportHandler) ExportSummary(c *gin.Context) {
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))
	group := strings.TrimSpace(c.DefaultQuery("group", "day"))
	storeID := strings.TrimSpace(c.Query("store_id"))
	method := strings.TrimSpace(c.Query("method"))
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "csv")))

	// 支持 day|store|method；其他回退 day
	if group != "day" && group != "store" && group != "method" {
		group = "day"
	}

	// 复用 Summary 逻辑
	// 为避免重复查询，直接在此查询并生成 rows
	db := database.GetDB()

	var pays []model.Payment
	pq := db.Model(&model.Payment{}).Preload("Order").Where("status = ?", 2)
	if start != "" {
		pq = pq.Where("created_at >= ?", start)
	}
	if end != "" {
		pq = pq.Where("created_at <= ?", end)
	}
	if storeID != "" {
		pq = pq.Joins("JOIN orders o ON o.id = payments.order_id").Where("o.store_id = ?", storeID)
	}
	if method != "" {
		pq = pq.Where("payment_method = ?", method)
	}
	if err := pq.Find(&pays).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	var refs []model.Refund
	rq := db.Model(&model.Refund{}).Preload("Order").Preload("Payment").Where("status = ?", 2)
	if start != "" {
		rq = rq.Where("created_at >= ?", start)
	}
	if end != "" {
		rq = rq.Where("created_at <= ?", end)
	}
	if storeID != "" {
		rq = rq.Joins("JOIN orders o ON o.id = refunds.order_id").Where("o.store_id = ?", storeID)
	}
	if method != "" {
		rq = rq.Joins("JOIN payments p ON p.id = refunds.payment_id").Where("p.payment_method = ?", method)
	}
	if err := rq.Find(&refs).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	filename := "finance_summary_" + time.Now().Format("20060102150405")
	if group == "day" {
		dayMap := map[string]*dailyRow{}
		dateKey := func(t time.Time) string { return t.Format("2006-01-02") }
		for _, p := range pays {
			d := dateKey(p.CreatedAt)
			row := dayMap[d]
			if row == nil {
				row = &dailyRow{Date: d, PayAmount: decimal.NewFromInt(0), RefundAmt: decimal.NewFromInt(0), NetAmount: decimal.NewFromInt(0)}
				dayMap[d] = row
			}
			row.PayCount++
			row.PayAmount = row.PayAmount.Add(p.Amount)
		}
		for _, r := range refs {
			d := dateKey(r.CreatedAt)
			row := dayMap[d]
			if row == nil {
				row = &dailyRow{Date: d, PayAmount: decimal.NewFromInt(0), RefundAmt: decimal.NewFromInt(0), NetAmount: decimal.NewFromInt(0)}
				dayMap[d] = row
			}
			row.RefundCount++
			row.RefundAmt = row.RefundAmt.Add(r.RefundAmount)
		}
		days := make([]string, 0, len(dayMap))
		for d := range dayMap {
			days = append(days, d)
		}
		sort.Strings(days)
		if format == "xlsx" {
			xf := excelize.NewFile()
			sheet := xf.GetSheetName(0)
			headers := []string{"Date", "Pay Count", "Pay Amount", "Refund Count", "Refund Amount", "Net Amount"}
			for i, hname := range headers {
				col, _ := excelize.ColumnNumberToName(i + 1)
				_ = xf.SetCellValue(sheet, col+"1", hname)
			}
			for idx, d := range days {
				r := dayMap[d]
				r.NetAmount = r.PayAmount.Sub(r.RefundAmt)
				row := []any{r.Date, r.PayCount, r.PayAmount.String(), r.RefundCount, r.RefundAmt.String(), r.NetAmount.String()}
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
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", "attachment; filename="+url.PathEscape(filename+".csv"))
		_, _ = c.Writer.WriteString("date,pay_count,pay_amount,refund_count,refund_amount,net_amount\n")
		for _, d := range days {
			r := dayMap[d]
			r.NetAmount = r.PayAmount.Sub(r.RefundAmt)
			line := fmt.Sprintf("%s,%d,%s,%d,%s,%s\n", d, r.PayCount, r.PayAmount.String(), r.RefundCount, r.RefundAmt.String(), r.NetAmount.String())
			_, _ = c.Writer.WriteString(line)
		}
		return
	}

	if group == "store" {
		type storeRow struct {
			StoreID                         uint
			PayCount                        int64
			PayAmount, RefundAmt, NetAmount decimal.Decimal
			RefundCount                     int64
		}
		m := map[uint]*storeRow{}
		for _, p := range pays {
			sid := p.Order.StoreID
			row := m[sid]
			if row == nil {
				row = &storeRow{StoreID: sid, PayAmount: decimal.NewFromInt(0), RefundAmt: decimal.NewFromInt(0)}
				m[sid] = row
			}
			row.PayCount++
			row.PayAmount = row.PayAmount.Add(p.Amount)
		}
		for _, r := range refs {
			sid := r.Order.StoreID
			row := m[sid]
			if row == nil {
				row = &storeRow{StoreID: sid, PayAmount: decimal.NewFromInt(0), RefundAmt: decimal.NewFromInt(0)}
				m[sid] = row
			}
			row.RefundCount++
			row.RefundAmt = row.RefundAmt.Add(r.RefundAmount)
		}
		keys := make([]uint, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		var stores []model.Store
		if len(keys) > 0 {
			if err := db.Model(&model.Store{}).Where("id IN ?", keys).Find(&stores).Error; err != nil {
				utils.Error(c, utils.CodeError, err.Error())
				return
			}
		}
		sname := map[uint]string{}
		for _, s := range stores {
			sname[s.ID] = s.Name
		}
		if format == "xlsx" {
			xf := excelize.NewFile()
			sheet := xf.GetSheetName(0)
			headers := []string{"Store ID", "Store Name", "Pay Count", "Pay Amount", "Refund Count", "Refund Amount", "Net Amount"}
			for i, h := range headers {
				col, _ := excelize.ColumnNumberToName(i + 1)
				_ = xf.SetCellValue(sheet, col+"1", h)
			}
			for idx, k := range keys {
				r := m[k]
				r.NetAmount = r.PayAmount.Sub(r.RefundAmt)
				row := []any{k, sname[k], r.PayCount, r.PayAmount.String(), r.RefundCount, r.RefundAmt.String(), r.NetAmount.String()}
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
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", "attachment; filename="+url.PathEscape(filename+".csv"))
		_, _ = c.Writer.WriteString("store_id,store_name,pay_count,pay_amount,refund_count,refund_amount,net_amount\n")
		for _, k := range keys {
			r := m[k]
			r.NetAmount = r.PayAmount.Sub(r.RefundAmt)
			line := fmt.Sprintf("%d,%s,%d,%s,%d,%s,%s\n", k, sname[k], r.PayCount, r.PayAmount.String(), r.RefundCount, r.RefundAmt.String(), r.NetAmount.String())
			_, _ = c.Writer.WriteString(line)
		}
		return
	}

	// group == method
	type methodRow struct {
		Method                          int
		PayCount                        int64
		PayAmount, RefundAmt, NetAmount decimal.Decimal
		RefundCount                     int64
	}
	m := map[int]*methodRow{}
	for _, p := range pays {
		md := p.PaymentMethod
		row := m[md]
		if row == nil {
			row = &methodRow{Method: md, PayAmount: decimal.NewFromInt(0), RefundAmt: decimal.NewFromInt(0)}
			m[md] = row
		}
		row.PayCount++
		row.PayAmount = row.PayAmount.Add(p.Amount)
	}
	for _, r := range refs {
		md := 0
		if r.Payment.ID != 0 {
			md = r.Payment.PaymentMethod
		}
		row := m[md]
		if row == nil {
			row = &methodRow{Method: md, PayAmount: decimal.NewFromInt(0), RefundAmt: decimal.NewFromInt(0)}
			m[md] = row
		}
		row.RefundCount++
		row.RefundAmt = row.RefundAmt.Add(r.RefundAmount)
	}
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	if format == "xlsx" {
		xf := excelize.NewFile()
		sheet := xf.GetSheetName(0)
		headers := []string{"Method", "Pay Count", "Pay Amount", "Refund Count", "Refund Amount", "Net Amount"}
		for i, h := range headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			_ = xf.SetCellValue(sheet, col+"1", h)
		}
		for idx, k := range keys {
			r := m[k]
			r.NetAmount = r.PayAmount.Sub(r.RefundAmt)
			row := []any{k, r.PayCount, r.PayAmount.String(), r.RefundCount, r.RefundAmt.String(), r.NetAmount.String()}
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
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+url.PathEscape(filename+".csv"))
	_, _ = c.Writer.WriteString("method,pay_count,pay_amount,refund_count,refund_amount,net_amount\n")
	for _, k := range keys {
		r := m[k]
		r.NetAmount = r.PayAmount.Sub(r.RefundAmt)
		line := fmt.Sprintf("%d,%d,%s,%d,%s,%s\n", k, r.PayCount, r.PayAmount.String(), r.RefundCount, r.RefundAmt.String(), r.NetAmount.String())
		_, _ = c.Writer.WriteString(line)
	}
}

// GET /api/v1/admin/finance/reconcile/diff
// 根据支付记录与订单应付金额进行核对，返回存在差异的订单
func (h *FinanceReportHandler) ReconcileDiff(c *gin.Context) {
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))
	storeID := strings.TrimSpace(c.Query("store_id"))
	method := strings.TrimSpace(c.Query("method"))
	pstr := strings.TrimSpace(c.DefaultQuery("page", "1"))
	lstr := strings.TrimSpace(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(pstr)
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(lstr)
	if limit <= 0 || limit > 200 {
		limit = 20
	}

	db := database.GetDB()

	type payAgg struct {
		OrderID uint
		Sum     decimal.Decimal
	}
	var aggs []payAgg
	pq := db.Model(&model.Payment{}).Select("order_id, SUM(amount) as sum").Where("status = ?", 2)
	if start != "" {
		pq = pq.Where("created_at >= ?", start)
	}
	if end != "" {
		pq = pq.Where("created_at <= ?", end)
	}
	if method != "" {
		pq = pq.Where("payment_method = ?", method)
	}
	if storeID != "" {
		pq = pq.Joins("JOIN orders o ON o.id = payments.order_id").Where("o.store_id = ?", storeID)
	}
	if err := pq.Group("order_id").Find(&aggs).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	if len(aggs) == 0 {
		utils.Success(c, gin.H{"total": 0, "rows": []gin.H{}})
		return
	}

	ids := make([]uint, 0, len(aggs))
	for _, a := range aggs {
		ids = append(ids, a.OrderID)
	}
	var orders []model.Order
	oq := db.Model(&model.Order{}).Where("id IN ?", ids)
	if err := oq.Find(&orders).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	omap := map[uint]model.Order{}
	for _, o := range orders {
		omap[o.ID] = o
	}

	type diff struct {
		OrderID  uint            `json:"order_id"`
		OrderNo  string          `json:"order_no"`
		StoreID  uint            `json:"store_id"`
		OrderPay decimal.Decimal `json:"order_pay_amount"`
		PaidSum  decimal.Decimal `json:"paid_amount_sum"`
		Diff     decimal.Decimal `json:"diff_amount"`
	}
	rows := make([]diff, 0)
	for _, a := range aggs {
		o := omap[a.OrderID]
		d := a.Sum.Sub(o.PayAmount)
		if !d.IsZero() {
			rows = append(rows, diff{OrderID: o.ID, OrderNo: o.OrderNo, StoreID: o.StoreID, OrderPay: o.PayAmount, PaidSum: a.Sum, Diff: d})
		}
	}

	total := len(rows)
	startIdx := (page - 1) * limit
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + limit
	if endIdx > total {
		endIdx = total
	}
	if startIdx > endIdx {
		startIdx = endIdx
	}
	utils.Success(c, gin.H{"total": total, "rows": rows[startIdx:endIdx]})
}

// GET /api/v1/admin/finance/reconcile/diff/export?format=csv|xlsx
func (h *FinanceReportHandler) ExportReconcileDiff(c *gin.Context) {
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))
	storeID := strings.TrimSpace(c.Query("store_id"))
	method := strings.TrimSpace(c.Query("method"))
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "csv")))

	db := database.GetDB()
	type payAgg struct {
		OrderID uint
		Sum     decimal.Decimal
	}
	var aggs []payAgg
	pq := db.Model(&model.Payment{}).Select("order_id, SUM(amount) as sum").Where("status = ?", 2)
	if start != "" {
		pq = pq.Where("created_at >= ?", start)
	}
	if end != "" {
		pq = pq.Where("created_at <= ?", end)
	}
	if method != "" {
		pq = pq.Where("payment_method = ?", method)
	}
	if storeID != "" {
		pq = pq.Joins("JOIN orders o ON o.id = payments.order_id").Where("o.store_id = ?", storeID)
	}
	if err := pq.Group("order_id").Find(&aggs).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	if len(aggs) == 0 {
		// 输出空模板
		if format == "xlsx" {
			xf := excelize.NewFile()
			sheet := xf.GetSheetName(0)
			headers := []string{"Order ID", "Order No", "Store ID", "Order Pay Amount", "Paid Amount Sum", "Diff Amount"}
			for i, h := range headers {
				col, _ := excelize.ColumnNumberToName(i + 1)
				_ = xf.SetCellValue(sheet, col+"1", h)
			}
			var buf bytes.Buffer
			if err := xf.Write(&buf); err != nil {
				utils.Error(c, utils.CodeError, err.Error())
				return
			}
			c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			c.Header("Content-Disposition", "attachment; filename=reconcile_diff.xlsx")
			_, _ = c.Writer.Write(buf.Bytes())
			return
		}
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", "attachment; filename=reconcile_diff.csv")
		_, _ = c.Writer.WriteString("order_id,order_no,store_id,order_pay_amount,paid_amount_sum,diff_amount\n")
		return
	}

	ids := make([]uint, 0, len(aggs))
	for _, a := range aggs {
		ids = append(ids, a.OrderID)
	}
	var orders []model.Order
	oq := db.Model(&model.Order{}).Where("id IN ?", ids)
	if err := oq.Find(&orders).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	omap := map[uint]model.Order{}
	for _, o := range orders {
		omap[o.ID] = o
	}

	type diff struct {
		OrderID                 uint
		OrderNo                 string
		StoreID                 uint
		OrderPay, PaidSum, Diff decimal.Decimal
	}
	rows := make([]diff, 0)
	for _, a := range aggs {
		o := omap[a.OrderID]
		d := a.Sum.Sub(o.PayAmount)
		if !d.IsZero() {
			rows = append(rows, diff{OrderID: o.ID, OrderNo: o.OrderNo, StoreID: o.StoreID, OrderPay: o.PayAmount, PaidSum: a.Sum, Diff: d})
		}
	}

	filename := "reconcile_diff_" + time.Now().Format("20060102150405")
	if format == "xlsx" {
		xf := excelize.NewFile()
		sheet := xf.GetSheetName(0)
		headers := []string{"Order ID", "Order No", "Store ID", "Order Pay Amount", "Paid Amount Sum", "Diff Amount"}
		for i, h := range headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			_ = xf.SetCellValue(sheet, col+"1", h)
		}
		for idx, r := range rows {
			row := []any{r.OrderID, r.OrderNo, r.StoreID, r.OrderPay.String(), r.PaidSum.String(), r.Diff.String()}
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
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+url.PathEscape(filename+".csv"))
	_, _ = c.Writer.WriteString("order_id,order_no,store_id,order_pay_amount,paid_amount_sum,diff_amount\n")
	for _, r := range rows {
		line := fmt.Sprintf("%d,%s,%d,%s,%s,%s\n", r.OrderID, r.OrderNo, r.StoreID, r.OrderPay.String(), r.PaidSum.String(), r.Diff.String())
		_, _ = c.Writer.WriteString(line)
	}
}
