package handler

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"tea-api/internal/model"
	"tea-api/internal/service/commission"
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
	if rec.Status != model.WithdrawStatusPending { // 仅申请中可受理
		utils.Error(c, utils.CodeError, "当前状态不可受理")
		return
	}
	uid, _ := c.Get("user_id")
	now := time.Now()
	rec.Status = model.WithdrawStatusProcessing
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
	if rec.Status != model.WithdrawStatusProcessing { // 仅处理中可完成
		utils.Error(c, utils.CodeError, "当前状态不可完成")
		return
	}
	uid, _ := c.Get("user_id")
	now := time.Now()
	rec.Status = model.WithdrawStatusCompleted
	rec.ProcessedAt = &now
	if u, ok := uid.(uint); ok {
		rec.ProcessedBy = u
	}
	// 统一完成阶段的备注为 JSON（含 withdraw_no/amount/fee/net）
	toCents := func(d decimal.Decimal) int64 { return d.Mul(decimal.NewFromInt(100)).IntPart() }
	rmkPaid := buildPaidRemark(rec.WithdrawNo, toCents(rec.Amount), toCents(rec.Fee), toCents(rec.ActualAmount))
	rec.Remark = rmkPaid
	if err := db.Save(&rec).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	// paid：解冻并完成扣减（申请时已将可用余额转入冻结，这里仅减少冻结并记录完成流水）
	if err := finalizeWalletForWithdraw(db, &rec); err != nil {
		utils.Error(c, utils.CodeError, fmt.Sprintf("wallet deduction failed: %s", err.Error()))
		return
	}

	// 按时间顺序消费该用户已解冻佣金，直至覆盖本次实际打款金额
	var opIDPtr *uint
	if u, ok := uid.(uint); ok && u != 0 {
		opIDPtr = &u
	}
	note := fmt.Sprintf("withdraw %s", rec.WithdrawNo)
	if _, consumed, err := commission.ConsumeUserAvailableCommissions(rec.UserID, rec.ActualAmount, rec.WithdrawNo, opIDPtr, note); err != nil {
		// 若佣金不足以完全覆盖本次提现，返回明确错误信息
		utils.Error(c, utils.CodeError, fmt.Sprintf("insufficient commission: consumed=%s, required=%s, err=%s", consumed.String(), rec.ActualAmount.String(), err.Error()))
		return
	}
	utils.Success(c, rec)
}

// finalizeWalletForWithdraw 解冻冻结金额并记录完成流水（金额单位：分）
func finalizeWalletForWithdraw(db *gorm.DB, rec *model.WithdrawRecord) error {
	// 将 decimal 金额转换为分（int64）
	toCents := func(d decimal.Decimal) int64 { return d.Mul(decimal.NewFromInt(100)).IntPart() }
	amountCents := toCents(rec.Amount)
	feeCents := toCents(rec.Fee)
	netCents := toCents(rec.ActualAmount)

	// 开启事务
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 读取并锁定钱包当前冻结金额
	var row struct{ Frozen int64 }
	if err := tx.Table("wallets").Select("frozen").Where("user_id = ?", rec.UserID).Take(&row).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("query wallet frozen: %w", err)
	}
	if row.Frozen < amountCents {
		tx.Rollback()
		return fmt.Errorf("insufficient wallet frozen: have=%d need=%d", row.Frozen, amountCents)
	}
	newFrozen := row.Frozen - amountCents
	if err := tx.Exec("UPDATE wallets SET frozen = ? WHERE user_id = ?", newFrozen, rec.UserID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("update wallet frozen: %w", err)
	}
	// 记录完成流水（对可用余额无影响，amount=0；备注体现手续费与实付）
	// 为便于对账，可查询最新余额作为 balance_after
	var balRow struct{ Balance int64 }
	if err := tx.Table("wallets").Select("balance").Where("user_id = ?", rec.UserID).Take(&balRow).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("query wallet balance: %w", err)
	}
	// remark 统一为 JSON
	rmkJSON := buildPaidRemark(rec.WithdrawNo, amountCents, feeCents, netCents)
	if err := tx.Exec("INSERT INTO wallet_transactions (user_id, type, amount, balance_after, remark, created_at) VALUES (?,?,?,?,?,NOW())",
		rec.UserID, "withdraw_paid", 0, balRow.Balance, rmkJSON).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("insert wallet tx: %w", err)
	}
	return tx.Commit().Error
}

// rollbackWalletForWithdrawReject 拒绝提现：从冻结中释放并返还到可用余额；记录 JSON remark
func rollbackWalletForWithdrawReject(db *gorm.DB, rec *model.WithdrawRecord) error {
	// 将 decimal 金额转换为分（int64）
	toCents := func(d decimal.Decimal) int64 { return d.Mul(decimal.NewFromInt(100)).IntPart() }
	amountCents := toCents(rec.Amount)
	feeCents := toCents(rec.Fee)
	netCents := toCents(rec.ActualAmount)

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	// 读取并锁定钱包当前余额与冻结
	var row struct {
		Balance int64
		Frozen  int64
	}
	if err := tx.Table("wallets").Select("balance,frozen").Where("user_id = ?", rec.UserID).Take(&row).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("query wallet: %w", err)
	}
	if row.Frozen < amountCents {
		tx.Rollback()
		return fmt.Errorf("insufficient frozen to rollback: have=%d need=%d", row.Frozen, amountCents)
	}
	newFrozen := row.Frozen - amountCents
	newBalance := row.Balance + amountCents
	if err := tx.Exec("UPDATE wallets SET balance = ?, frozen = ? WHERE user_id = ?", newBalance, newFrozen, rec.UserID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("update wallet rollback: %w", err)
	}
	// 插入解冻流水（可用余额增加），remark 为 JSON
	rmkJSON := buildRejectUnfreezeRemark(rec.WithdrawNo, amountCents, feeCents, netCents)
	if err := tx.Exec("INSERT INTO wallet_transactions (user_id, type, amount, balance_after, remark, created_at) VALUES (?,?,?,?,?,NOW())",
		rec.UserID, "withdraw_reject_unfreeze", amountCents, newBalance, rmkJSON).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("insert wallet tx rollback: %w", err)
	}
	return tx.Commit().Error
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
	if rec.Status != model.WithdrawStatusPending && rec.Status != model.WithdrawStatusProcessing { // 申请中/处理中可拒绝
		utils.Error(c, utils.CodeError, "当前状态不可拒绝")
		return
	}
	uid, _ := c.Get("user_id")
	now := time.Now()
	rec.Status = model.WithdrawStatusRejected
	rec.ProcessedAt = &now
	if u, ok := uid.(uint); ok {
		rec.ProcessedBy = u
	}
	// 统一拒绝阶段的备注为 JSON（含 withdraw_no/amount/fee/net）
	toCents := func(d decimal.Decimal) int64 { return d.Mul(decimal.NewFromInt(100)).IntPart() }
	rmkRej := buildRejectUnfreezeRemark(rec.WithdrawNo, toCents(rec.Amount), toCents(rec.Fee), toCents(rec.ActualAmount))
	rec.Remark = rmkRej
	if err := db.Save(&rec).Error; err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}
	// 拒绝：解冻并恢复可用余额，记录解冻流水
	if err := rollbackWalletForWithdrawReject(db, &rec); err != nil {
		utils.Error(c, utils.CodeError, fmt.Sprintf("wallet rollback failed: %s", err.Error()))
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
