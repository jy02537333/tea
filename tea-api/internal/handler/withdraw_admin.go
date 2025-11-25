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

type WithdrawAdminHandler struct{}

func NewWithdrawAdminHandler() *WithdrawAdminHandler { return &WithdrawAdminHandler{} }

type withdrawActionReq struct {
	Remark string `json:"remark"`
}

// GET /api/v1/admin/withdraws
// 查询参数：user_id, withdraw_no(模糊), status, start, end, page, limit
func (h *WithdrawAdminHandler) List(c *gin.Context) {
	userID := strings.TrimSpace(c.Query("user_id"))
	wno := strings.TrimSpace(c.Query("withdraw_no"))
	status := strings.TrimSpace(c.Query("status"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))

	page := toIntWd(c.DefaultQuery("page", "1"))
	size := toIntWd(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}

	db := database.GetDB()
	q := db.Model(&model.WithdrawRecord{}).Preload("User")
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	if wno != "" {
		q = q.Where("withdraw_no LIKE ?", "%"+wno+"%")
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

	var list []model.WithdrawRecord
	if err := q.Order("id desc").Limit(size).Offset((page - 1) * size).Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.PageSuccess(c, list, total, page, size)
}

// GET /api/v1/admin/withdraws/export?format=csv|xlsx
// 导出最近最多5000条，支持与 List 相同的过滤参数
func (h *WithdrawAdminHandler) Export(c *gin.Context) {
	userID := strings.TrimSpace(c.Query("user_id"))
	wno := strings.TrimSpace(c.Query("withdraw_no"))
	status := strings.TrimSpace(c.Query("status"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "csv")))

	db := database.GetDB()
	q := db.Model(&model.WithdrawRecord{}).Preload("User")
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	if wno != "" {
		q = q.Where("withdraw_no LIKE ?", "%"+wno+"%")
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

	var list []model.WithdrawRecord
	if err := q.Order("id desc").Limit(5000).Find(&list).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	filename := "withdraws_" + time.Now().Format("20060102150405")
	if format == "xlsx" {
		xf := excelize.NewFile()
		sheet := xf.GetSheetName(0)
		headers := []string{"ID", "Withdraw No", "User ID", "Amount", "Fee", "Actual Amount", "Type", "Status", "Processed At", "Processed By", "Created At"}
		for i, h := range headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			_ = xf.SetCellValue(sheet, col+"1", h)
		}
		for idx, it := range list {
			procAt := ""
			if it.ProcessedAt != nil {
				procAt = it.ProcessedAt.Format("2006-01-02 15:04:05")
			}
			row := []any{it.ID, it.WithdrawNo, it.UserID, it.Amount.String(), it.Fee.String(), it.ActualAmount.String(), it.WithdrawType, it.Status, procAt, it.ProcessedBy, it.CreatedAt.Format("2006-01-02 15:04:05")}
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

	// CSV 默认
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+url.PathEscape(filename+".csv"))
	_, _ = c.Writer.WriteString("id,withdraw_no,user_id,amount,fee,actual_amount,type,status,processed_at,processed_by,created_at\n")
	for _, it := range list {
		procAt := ""
		if it.ProcessedAt != nil {
			procAt = it.ProcessedAt.Format("2006-01-02 15:04:05")
		}
		line := fmt.Sprintf("%d,%s,%d,%s,%s,%s,%d,%d,%s,%d,%s\n", it.ID, csvSafeWd(it.WithdrawNo), it.UserID, csvSafeWd(it.Amount.String()), csvSafeWd(it.Fee.String()), csvSafeWd(it.ActualAmount.String()), it.WithdrawType, it.Status, csvSafeWd(procAt), it.ProcessedBy, it.CreatedAt.Format("2006-01-02 15:04:05"))
		_, _ = c.Writer.WriteString(line)
	}
}

// POST /api/v1/admin/withdraws/:id/approve  将状态置为处理中(2)
func (h *WithdrawAdminHandler) Approve(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req withdrawActionReq
	_ = c.ShouldBindJSON(&req)
	db := database.GetDB()
	var rec model.WithdrawRecord
	if err := db.First(&rec, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "记录不存在")
		return
	}
	if rec.Status != 1 { // 仅申请中可受理
		utils.Error(c, utils.CodeError, "当前状态不可受理")
		return
	}
	uid, _ := c.Get("user_id")
	now := time.Now()
	rec.Status = 2
	rec.ProcessedAt = &now
	if u, ok := uid.(uint); ok {
		rec.ProcessedBy = u
	}
	if req.Remark != "" {
		rec.Remark = req.Remark
	}
	if err := db.Save(&rec).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, rec)
}

// POST /api/v1/admin/withdraws/:id/complete 将状态置为已完成(3)
func (h *WithdrawAdminHandler) Complete(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req withdrawActionReq
	_ = c.ShouldBindJSON(&req)
	db := database.GetDB()
	var rec model.WithdrawRecord
	if err := db.First(&rec, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "记录不存在")
		return
	}
	if rec.Status != 2 { // 仅处理中可完成
		utils.Error(c, utils.CodeError, "当前状态不可完成")
		return
	}
	uid, _ := c.Get("user_id")
	now := time.Now()
	rec.Status = 3
	rec.ProcessedAt = &now
	if u, ok := uid.(uint); ok {
		rec.ProcessedBy = u
	}
	if req.Remark != "" {
		rec.Remark = req.Remark
	}
	if err := db.Save(&rec).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, rec)
}

// POST /api/v1/admin/withdraws/:id/reject 将状态置为已拒绝(4)
func (h *WithdrawAdminHandler) Reject(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req withdrawActionReq
	_ = c.ShouldBindJSON(&req)
	db := database.GetDB()
	var rec model.WithdrawRecord
	if err := db.First(&rec, id).Error; err != nil {
		utils.Error(c, utils.CodeError, "记录不存在")
		return
	}
	if rec.Status != 1 && rec.Status != 2 { // 申请中/处理中可拒绝
		utils.Error(c, utils.CodeError, "当前状态不可拒绝")
		return
	}
	uid, _ := c.Get("user_id")
	now := time.Now()
	rec.Status = 4
	rec.ProcessedAt = &now
	if u, ok := uid.(uint); ok {
		rec.ProcessedBy = u
	}
	if req.Remark != "" {
		rec.Remark = req.Remark
	}
	if err := db.Save(&rec).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	utils.Success(c, rec)
}

func toIntWd(s string) int { var n int; _, _ = fmt.Sscanf(s, "%d", &n); return n }
func csvSafeWd(s string) string {
	s = strings.ReplaceAll(s, ",", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
