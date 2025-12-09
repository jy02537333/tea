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

type LogsHandler struct{}

func NewLogsHandler() *LogsHandler { return &LogsHandler{} }

// GET /api/v1/admin/logs/operations
func (h *LogsHandler) ListOperationLogs(c *gin.Context) {
	module := strings.TrimSpace(c.Query("module"))
	method := strings.TrimSpace(c.Query("method"))
	operation := strings.TrimSpace(c.Query("operation"))
	orderID := strings.TrimSpace(c.Query("order_id"))
	pathLike := strings.TrimSpace(c.Query("path"))
	userID := strings.TrimSpace(c.Query("user_id"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))

	page := toInt(c.DefaultQuery("page", "1"))
	size := toInt(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}

	db := database.GetDB()
	q := db.Model(&model.OperationLog{})
	if module != "" {
		q = q.Where("module = ?", module)
	}
	if method != "" {
		q = q.Where("operation LIKE ?", method+" %")
	}
	if operation != "" {
		q = q.Where("operation = ?", operation)
	}
	if pathLike != "" {
		q = q.Where("operation LIKE ?", "% "+pathLike+"%")
	}
	if orderID != "" {
		// 过滤 RequestData JSON 中的 order_id（json.Marshal 默认无空格）
		q = q.Where("request_data LIKE ?", "%\"order_id\":"+orderID+"%")
	}
	if userID != "" {
		q = q.Where("user_id = ?", userID)
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
	var list []model.OperationLog
	if err := q.Order("id desc").Limit(size).Offset((page - 1) * size).Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.PageSuccess(c, list, total, page, size)
}

// GET /api/v1/admin/logs/operations/export?format=csv
func (h *LogsHandler) ExportOperationLogs(c *gin.Context) {
	module := strings.TrimSpace(c.Query("module"))
	method := strings.TrimSpace(c.Query("method"))
	operation := strings.TrimSpace(c.Query("operation"))
	orderID := strings.TrimSpace(c.Query("order_id"))
	pathLike := strings.TrimSpace(c.Query("path"))
	userID := strings.TrimSpace(c.Query("user_id"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "csv")))

	db := database.GetDB()
	q := db.Model(&model.OperationLog{})
	if module != "" {
		q = q.Where("module = ?", module)
	}
	if method != "" {
		q = q.Where("operation LIKE ?", method+" %")
	}
	if operation != "" {
		q = q.Where("operation = ?", operation)
	}
	if pathLike != "" {
		q = q.Where("operation LIKE ?", "% "+pathLike+"%")
	}
	if orderID != "" {
		q = q.Where("request_data LIKE ?", "%\"order_id\":"+orderID+"%")
	}
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	if start != "" {
		q = q.Where("created_at >= ?", start)
	}
	if end != "" {
		q = q.Where("created_at <= ?", end)
	}

	var list []model.OperationLog
	if err := q.Order("id desc").Limit(5000).Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	filename := "operation_logs_" + time.Now().Format("20060102150405")
	if format == "xlsx" {
		xf := excelize.NewFile()
		sheet := xf.GetSheetName(0)
		headers := []string{"ID", "User ID", "Module", "Operation", "IP", "UA", "Created At"}
		for i, hname := range headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			_ = xf.SetCellValue(sheet, col+"1", hname)
		}
		for idx, it := range list {
			row := []any{
				it.ID,
				it.UserID,
				it.Module,
				it.Operation,
				it.IP,
				it.UserAgent,
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

	// header
	_, _ = c.Writer.WriteString("id,user_id,module,operation,ip,ua,created_at\n")
	for _, it := range list {
		line := fmt.Sprintf("%d,%d,%s,%s,%s,%s,%s\n",
			it.ID,
			it.UserID,
			csvSafe(it.Module),
			csvSafe(it.Operation),
			csvSafe(it.IP),
			csvSafe(it.UserAgent),
			it.CreatedAt.Format("2006-01-02 15:04:05"),
		)
		_, _ = c.Writer.WriteString(line)
	}
}

func toInt(s string) int { var n int; _, _ = fmt.Sscanf(s, "%d", &n); return n }
func csvSafe(s string) string {
	s = strings.ReplaceAll(s, ",", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// GET /api/v1/admin/logs/access
func (h *LogsHandler) ListAccessLogs(c *gin.Context) {
	method := strings.TrimSpace(c.Query("method"))
	pathLike := strings.TrimSpace(c.Query("path"))
	userID := strings.TrimSpace(c.Query("user_id"))
	status := strings.TrimSpace(c.Query("status"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))

	page := toInt(c.DefaultQuery("page", "1"))
	size := toInt(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}

	db := database.GetDB()
	q := db.Model(&model.AccessLog{})
	if method != "" {
		q = q.Where("method = ?", method)
	}
	if pathLike != "" {
		q = q.Where("path LIKE ?", "%"+pathLike+"%")
	}
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	if status != "" {
		q = q.Where("status_code = ?", status)
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
	var list []model.AccessLog
	if err := q.Order("id desc").Limit(size).Offset((page - 1) * size).Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.PageSuccess(c, list, total, page, size)
}

// GET /api/v1/admin/logs/access/export
func (h *LogsHandler) ExportAccessLogs(c *gin.Context) {
	method := strings.TrimSpace(c.Query("method"))
	pathLike := strings.TrimSpace(c.Query("path"))
	userID := strings.TrimSpace(c.Query("user_id"))
	status := strings.TrimSpace(c.Query("status"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "csv")))

	db := database.GetDB()
	q := db.Model(&model.AccessLog{})
	if method != "" {
		q = q.Where("method = ?", method)
	}
	if pathLike != "" {
		q = q.Where("path LIKE ?", "%"+pathLike+"%")
	}
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	if status != "" {
		q = q.Where("status_code = ?", status)
	}
	if start != "" {
		q = q.Where("created_at >= ?", start)
	}
	if end != "" {
		q = q.Where("created_at <= ?", end)
	}

	var list []model.AccessLog
	if err := q.Order("id desc").Limit(5000).Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	filename := "access_logs_" + time.Now().Format("20060102150405")
	if format == "xlsx" {
		xf := excelize.NewFile()
		sheet := xf.GetSheetName(0)
		headers := []string{"ID", "User ID", "Method", "Path", "Query", "Status", "Latency", "IP", "UA", "Created At"}
		for i, hname := range headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			_ = xf.SetCellValue(sheet, col+"1", hname)
		}
		for idx, it := range list {
			row := []any{
				it.ID,
				derefUint(it.UserID),
				it.Method,
				it.Path,
				it.Query,
				it.StatusCode,
				it.Latency,
				it.IP,
				it.UserAgent,
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
	_, _ = c.Writer.WriteString("id,user_id,method,path,query,status_code,latency,ip,ua,created_at\n")
	for _, it := range list {
		line := fmt.Sprintf("%d,%d,%s,%s,%s,%d,%d,%s,%s,%s\n",
			it.ID,
			derefUint(it.UserID),
			csvSafe(it.Method),
			csvSafe(it.Path),
			csvSafe(it.Query),
			it.StatusCode,
			it.Latency,
			csvSafe(it.IP),
			csvSafe(it.UserAgent),
			it.CreatedAt.Format("2006-01-02 15:04:05"),
		)
		_, _ = c.Writer.WriteString(line)
	}
}

func derefUint(val *uint) uint {
	if val == nil {
		return 0
	}
	return *val
}
