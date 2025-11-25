package handler

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/database"
	"tea-api/pkg/utils"
)

type AccrualHandler struct {
	svc *service.AccrualService
}

func NewAccrualHandler() *AccrualHandler {
	return &AccrualHandler{svc: service.NewAccrualService()}
}

// AdminAccrualRun 触发计息（管理员）
// POST /api/v1/admin/accrual/run
// body: {"date":"2025-11-12","rate":0.001, "user_id":123 (optional)}
func (h *AccrualHandler) AdminAccrualRun(c *gin.Context) {
	var req struct {
		Date   string  `json:"date"`
		Rate   float64 `json:"rate"`
		UserID *uint   `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.InvalidParam(c, err.Error())
		return
	}
	if req.Rate <= 0 {
		req.Rate = 0.001
	} // 默认万分之一/日
	var dt time.Time
	var err error
	if req.Date == "" {
		dt = time.Now()
	} else {
		dt, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			utils.InvalidParam(c, "date 格式应为 YYYY-MM-DD")
			return
		}
	}
	n, err := h.svc.Run(dt, req.Rate, req.UserID)
	if err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, gin.H{"updated": n})
}

// UserInterestRecords 当前用户的计息记录列表
// GET /api/v1/user/interest-records?page=1&limit=20
func (h *AccrualHandler) UserInterestRecords(c *gin.Context) {
	uidVal, ok := c.Get("user_id")
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}
	uid, _ := uidVal.(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}

	db := database.GetDB()
	var total int64
	if err := db.Model(&model.InterestRecord{}).Where("user_id = ?", uid).Count(&total).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	var list []model.InterestRecord
	if err := db.Where("user_id = ?", uid).Order("date desc, id desc").Limit(size).Offset((page - 1) * size).Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.PageSuccess(c, list, total, page, size)
}

// AdminAccrualExport 导出计息记录为CSV
// GET /api/v1/admin/accrual/export?start=YYYY-MM-DD&end=YYYY-MM-DD&user_id=&limit=
func (h *AccrualHandler) AdminAccrualExport(c *gin.Context) {
	start := c.Query("start")
	end := c.Query("end")
	userID := c.Query("user_id")
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "csv")))
	lang := strings.ToLower(strings.TrimSpace(c.DefaultQuery("lang", "zh")))
	fieldsParam := strings.TrimSpace(c.Query("fields"))
	wantZip := c.DefaultQuery("zip", "0") == "1"

	db := database.GetDB()
	q := db.Model(&model.InterestRecord{})
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	if start != "" {
		q = q.Where("date >= ?", start)
	}
	if end != "" {
		q = q.Where("date <= ?", end)
	}

	var list []model.InterestRecord
	if err := q.Order("date asc, id asc").Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	// 字段选择与表头多语言
	allFields := []string{"user_id", "date", "principal_before", "rate", "interest_amount", "principal_after", "method"}
	chosen := allFields
	if fieldsParam != "" {
		parts := strings.Split(fieldsParam, ",")
		tmp := make([]string, 0, len(parts))
		// 去空格 + 去重 + 合法过滤
		set := map[string]struct{}{}
		allowed := map[string]struct{}{}
		for _, f := range allFields {
			allowed[f] = struct{}{}
		}
		for _, p := range parts {
			f := strings.ToLower(strings.TrimSpace(p))
			if f == "" {
				continue
			}
			if _, ok := allowed[f]; !ok {
				continue
			}
			if _, exists := set[f]; exists {
				continue
			}
			set[f] = struct{}{}
			tmp = append(tmp, f)
		}
		if len(tmp) > 0 {
			chosen = tmp
		}
	}
	// 固定顺序，按 allFields 排序 chosen
	order := map[string]int{}
	for i, f := range allFields {
		order[f] = i
	}
	sort.SliceStable(chosen, func(i, j int) bool { return order[chosen[i]] < order[chosen[j]] })

	headerZH := map[string]string{
		"user_id": "用户ID", "date": "日期", "principal_before": "期初本金", "rate": "日利率", "interest_amount": "利息", "principal_after": "期末本金", "method": "方法",
	}
	headerEN := map[string]string{
		"user_id": "User ID", "date": "Date", "principal_before": "Principal Before", "rate": "Rate", "interest_amount": "Interest", "principal_after": "Principal After", "method": "Method",
	}
	header := func(key string) string {
		if lang == "en" {
			if v, ok := headerEN[key]; ok {
				return v
			}
		}
		if v, ok := headerZH[key]; ok {
			return v
		}
		return key
	}

	// 生成数据行的函数
	rowVals := func(r model.InterestRecord) []string {
		vals := make([]string, 0, len(chosen))
		for _, f := range chosen {
			switch f {
			case "user_id":
				vals = append(vals, fmt.Sprintf("%d", r.UserID))
			case "date":
				vals = append(vals, r.Date.Format("2006-01-02"))
			case "principal_before":
				vals = append(vals, r.PrincipalBefore.String())
			case "rate":
				vals = append(vals, r.Rate.String())
			case "interest_amount":
				vals = append(vals, r.InterestAmount.String())
			case "principal_after":
				vals = append(vals, r.PrincipalAfter.String())
			case "method":
				vals = append(vals, r.Method)
			default:
				vals = append(vals, "")
			}
		}
		return vals
	}

	filename := "interest_records"
	if format == "xlsx" {
		// 生成 XLSX
		xf := excelize.NewFile()
		sheet := xf.GetSheetName(0)
		// 写表头
		for i, f := range chosen {
			col, _ := excelize.ColumnNumberToName(i + 1)
			_ = xf.SetCellValue(sheet, col+"1", header(f))
		}
		// 写数据
		for idx, it := range list {
			for j, f := range rowVals(it) {
				col, _ := excelize.ColumnNumberToName(j + 1)
				_ = xf.SetCellValue(sheet, col+fmt.Sprintf("%d", idx+2), f)
			}
		}
		var buf bytes.Buffer
		if err := xf.Write(&buf); err != nil {
			utils.Error(c, utils.CodeError, err.Error())
			return
		}

		if wantZip {
			var zbuf bytes.Buffer
			zw := zip.NewWriter(&zbuf)
			f, _ := zw.Create(filename + ".xlsx")
			_, _ = f.Write(buf.Bytes())
			_ = zw.Close()

			c.Header("Content-Type", "application/zip")
			c.Header("Content-Disposition", "attachment; filename="+url.PathEscape(filename+".zip"))
			_, _ = c.Writer.Write(zbuf.Bytes())
			return
		}
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename="+url.PathEscape(filename+".xlsx"))
		_, _ = c.Writer.Write(buf.Bytes())
		return
	}

	// 默认 CSV 输出
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+url.PathEscape(filename+".csv"))
	// 写CSV头
	for i, f := range chosen {
		if i > 0 {
			_, _ = c.Writer.WriteString(",")
		}
		_, _ = c.Writer.WriteString(header(f))
	}
	_, _ = c.Writer.WriteString("\n")
	for _, r := range list {
		vals := rowVals(r)
		for i, v := range vals {
			if i > 0 {
				_, _ = c.Writer.WriteString(",")
			}
			// 简单转义逗号
			v = strings.ReplaceAll(v, ",", " ")
			_, _ = c.Writer.WriteString(v)
		}
		_, _ = c.Writer.WriteString("\n")
	}
}

// AdminAccrualSummary 汇总统计
// GET /api/v1/admin/accrual/summary?start=YYYY-MM-DD&end=YYYY-MM-DD
func (h *AccrualHandler) AdminAccrualSummary(c *gin.Context) {
	start := c.Query("start")
	end := c.Query("end")
	db := database.GetDB()
	type row struct {
		Count int64
		Users int64
		Sum   string
	}
	var res row
	q := db.Table("interest_records").Select("COUNT(*) as count, COUNT(DISTINCT user_id) as users, COALESCE(SUM(interest_amount),0) as sum")
	if start != "" {
		q = q.Where("date >= ?", start)
	}
	if end != "" {
		q = q.Where("date <= ?", end)
	}
	if err := q.Scan(&res).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, gin.H{
		"record_count":   res.Count,
		"user_count":     res.Users,
		"total_interest": res.Sum,
	})
}
