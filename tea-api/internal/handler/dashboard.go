package handler

import (
	"net/http"
	"strconv"
	"time"

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

// Summary 返回后台首页数据摘要（最小版）
// GET /api/v1/admin/dashboard/summary?range=today
// 返回字段：today_order_count, today_sales_amount, today_paid_order_count
func (h *DashboardHandler) Summary(c *gin.Context) {
	db := database.GetDB()

	// 计算当天时间范围（本地时区）
	// 最小版：使用数据库的 DATE(created_at) = CURDATE() 近似统计
	type row struct {
		Cnt int64
	}

	var totalToday int64
	if err := db.Model(&model.Order{}).
		Where("DATE(created_at) = CURDATE()").
		Count(&totalToday).Error; err != nil {
		// 容错：缺表时返回0而非500
		c.JSON(http.StatusOK, gin.H{"data": gin.H{
			"today_order_count":      0,
			"today_sales_amount":     0,
			"today_paid_order_count": 0,
		}})
		return
	}

	// 已付款订单数量（今日）
	var paidToday int64
	_ = db.Model(&model.Order{}).
		Where("DATE(created_at) = CURDATE() AND pay_status = ?", 2).
		Count(&paidToday).Error

	// 今日销售额：sum(pay_amount) where pay_status=2 and DATE(paid_at)=CURDATE()
	// 若 paid_at 为空则按 created_at 近似
	type sumRow struct {
		Sum float64
	}
	var s sumRow
	_ = db.Model(&model.Order{}).
		Select("COALESCE(SUM(pay_amount),0) as sum").
		Where("pay_status = 2 AND (DATE(paid_at) = CURDATE() OR (paid_at IS NULL AND DATE(created_at) = CURDATE()))").
		Scan(&s).Error

	// 今日退款额：Refund.status=2 且 DATE(refunded_at)=CURDATE()
	var todayRefund struct{ Sum float64 }
	_ = db.Model(&model.Refund{}).
		Select("COALESCE(SUM(refund_amount),0) as sum").
		Where("status = 2 AND refunded_at IS NOT NULL AND DATE(refunded_at) = CURDATE()").
		Scan(&todayRefund).Error

	// 昨日与近7/30日聚合（最小实现，避免复杂 SQL，基于日期条件）
	// 昨日：DATE(created_at) = CURDATE() - INTERVAL 1 DAY
	var yPaid int64
	_ = db.Model(&model.Order{}).
		Where("pay_status = 2 AND DATE(created_at) = (CURDATE() - INTERVAL 1 DAY)").
		Count(&yPaid).Error
	type sumRow2 struct{ Sum float64 }
	var ySales sumRow2
	_ = db.Model(&model.Order{}).
		Select("COALESCE(SUM(pay_amount),0) as sum").
		Where("pay_status = 2 AND (DATE(paid_at) = (CURDATE() - INTERVAL 1 DAY) OR (paid_at IS NULL AND DATE(created_at) = (CURDATE() - INTERVAL 1 DAY)))").
		Scan(&ySales).Error

	// 近7日：DATE(created_at) >= CURDATE() - INTERVAL 6 DAY
	var wPaid int64
	_ = db.Model(&model.Order{}).
		Where("pay_status = 2 AND DATE(created_at) >= (CURDATE() - INTERVAL 6 DAY)").
		Count(&wPaid).Error
	var wSales sumRow2
	_ = db.Model(&model.Order{}).
		Select("COALESCE(SUM(pay_amount),0) as sum").
		Where("pay_status = 2 AND (DATE(paid_at) >= (CURDATE() - INTERVAL 6 DAY) OR (paid_at IS NULL AND DATE(created_at) >= (CURDATE() - INTERVAL 6 DAY)))").
		Scan(&wSales).Error

	// 近30日：DATE(created_at) >= CURDATE() - INTERVAL 29 DAY
	var mPaid int64
	_ = db.Model(&model.Order{}).
		Where("pay_status = 2 AND DATE(created_at) >= (CURDATE() - INTERVAL 29 DAY)").
		Count(&mPaid).Error
	var mSales sumRow2
	_ = db.Model(&model.Order{}).
		Select("COALESCE(SUM(pay_amount),0) as sum").
		Where("pay_status = 2 AND (DATE(paid_at) >= (CURDATE() - INTERVAL 29 DAY) OR (paid_at IS NULL AND DATE(created_at) >= (CURDATE() - INTERVAL 29 DAY)))").
		Scan(&mSales).Error

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"today_order_count":          totalToday,
		"today_sales_amount":         s.Sum,
		"today_paid_order_count":     paidToday,
		"today_refund_amount":        todayRefund.Sum,
		"yesterday_sales_amount":     ySales.Sum,
		"yesterday_paid_order_count": yPaid,
		"yesterday_refund_amount": func() float64 {
			var r2 sumRow2
			_ = db.Model(&model.Refund{}).
				Select("COALESCE(SUM(refund_amount),0) as sum").
				Where("status = 2 AND refunded_at IS NOT NULL AND DATE(refunded_at) = (CURDATE() - INTERVAL 1 DAY)").
				Scan(&r2).Error
			return r2.Sum
		}(),
		"last7d_sales_amount":     wSales.Sum,
		"last7d_paid_order_count": wPaid,
		"last7d_refund_amount": func() float64 {
			var r3 sumRow2
			_ = db.Model(&model.Refund{}).
				Select("COALESCE(SUM(refund_amount),0) as sum").
				Where("status = 2 AND refunded_at IS NOT NULL AND DATE(refunded_at) >= (CURDATE() - INTERVAL 6 DAY)").
				Scan(&r3).Error
			return r3.Sum
		}(),
		"last30d_sales_amount":     mSales.Sum,
		"last30d_paid_order_count": mPaid,
		"last30d_refund_amount": func() float64 {
			var r4 sumRow2
			_ = db.Model(&model.Refund{}).
				Select("COALESCE(SUM(refund_amount),0) as sum").
				Where("status = 2 AND refunded_at IS NOT NULL AND DATE(refunded_at) >= (CURDATE() - INTERVAL 29 DAY)").
				Scan(&r4).Error
			return r4.Sum
		}(),
	}})
}

// OrderTrends 返回订单趋势数据（最小版）
// GET /api/v1/admin/dashboard/order-trends?days=7
// 默认返回近 7 天（含今日）的逐日数据：date, order_count, paid_order_count, sales_amount
func (h *DashboardHandler) OrderTrends(c *gin.Context) {
	db := database.GetDB()

	// 解析 days 参数，默认 7，最大 60 以避免过大查询
	days := 7
	if v := c.Query("days"); v != "" {
		if d, err := strconv.Atoi(v); err == nil && d > 0 {
			if d > 60 {
				d = 60
			}
			days = d
		}
	}

	// 计算起始日期（含当天），如 days=7 则起点为今日-6天 00:00:00
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(days - 1))

	// 聚合查询：按天统计订单数、已支付订单数与销售额
	type row struct {
		D           string  `gorm:"column:d"`
		OrderCount  int64   `gorm:"column:order_count"`
		PaidCount   int64   `gorm:"column:paid_count"`
		SalesAmount float64 `gorm:"column:sales_amount"`
	}
	var rows []row
	if err := db.Model(&model.Order{}).
		Select("DATE(created_at) as d, COUNT(*) as order_count, SUM(CASE WHEN pay_status = 2 THEN 1 ELSE 0 END) as paid_count, COALESCE(SUM(CASE WHEN pay_status = 2 THEN pay_amount END),0) as sales_amount").
		Where("created_at >= ?", startDate).
		Group("DATE(created_at)").
		Order("DATE(created_at)").
		Scan(&rows).Error; err != nil {
		// 缺表等错误时，返回空数组以保证后台页面可用
		c.JSON(http.StatusOK, gin.H{"data": []any{}})
		return
	}

	// 组装连续日期序列并填充缺失日期为 0
	type point struct {
		Date           string  `json:"date"`
		OrderCount     int64   `json:"order_count"`
		PaidOrderCount int64   `json:"paid_order_count"`
		SalesAmount    float64 `json:"sales_amount"`
	}
	mp := make(map[string]row, len(rows))
	for _, r := range rows {
		mp[r.D] = r
	}
	result := make([]point, 0, days)
	for i := 0; i < days; i++ {
		d := startDate.AddDate(0, 0, i)
		key := d.Format("2006-01-02")
		if r, ok := mp[key]; ok {
			result = append(result, point{Date: key, OrderCount: r.OrderCount, PaidOrderCount: r.PaidCount, SalesAmount: r.SalesAmount})
		} else {
			result = append(result, point{Date: key, OrderCount: 0, PaidOrderCount: 0, SalesAmount: 0})
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}
