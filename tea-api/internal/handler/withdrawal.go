package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"tea-api/internal/config"
	"tea-api/pkg/database"
	"tea-api/pkg/response"
)

// ===== 用户银行卡/提现账户 =====

type bankAccountRow struct {
	ID          uint   `json:"id"`
	AccountType string `json:"type"`
	AccountName string `json:"account_name"`
	AccountNo   string `json:"masked_account"`
	BankName    string `json:"bank_name"`
	IsDefault   int    `json:"is_default"`
	CreatedAt   string `json:"created_at"`
}

// ListMyBankAccounts GET /api/v1/wallet/bank-accounts
func ListMyBankAccounts(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未获取到用户身份")
		return
	}
	var rows []bankAccountRow
	// 脱敏：仅展示尾号（假设 account_no 已存明文，生产建议加密存储）
	type rawRow struct {
		ID          uint
		AccountType string
		AccountName string
		AccountNo   string
		BankName    string
		IsDefault   int
		CreatedAt   time.Time
	}
	if err := database.GetDB().Table("user_bank_accounts").
		Select("id, account_type, account_name, account_no, bank_name, is_default, created_at").
		Where("user_id = ?", uid).Order("id desc").
		Find(&[]rawRow{}).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "查询提现账户失败")
		return
	}
	var raw []rawRow
	database.GetDB().Table("user_bank_accounts").
		Select("id, account_type, account_name, account_no, bank_name, is_default, created_at").
		Where("user_id = ?", uid).Order("id desc").Find(&raw)
	rows = make([]bankAccountRow, 0, len(raw))
	for _, r := range raw {
		masked := maskAccount(r.AccountNo)
		rows = append(rows, bankAccountRow{
			ID:          r.ID,
			AccountType: r.AccountType,
			AccountName: r.AccountName,
			AccountNo:   masked,
			BankName:    r.BankName,
			IsDefault:   r.IsDefault,
			CreatedAt:   r.CreatedAt.Format(time.RFC3339),
		})
	}
	response.Success(c, rows)
}

// CreateMyBankAccount POST /api/v1/wallet/bank-accounts
func CreateMyBankAccount(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未获取到用户身份")
		return
	}
	var req struct {
		Type        string `json:"type" binding:"required"`
		AccountName string `json:"account_name" binding:"required"`
		AccountNo   string `json:"account_no" binding:"required"`
		BankName    string `json:"bank_name"`
		IsDefault   int    `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Type == "" || req.AccountName == "" || req.AccountNo == "" {
		response.BadRequest(c, "参数错误")
		return
	}
	tx := database.GetDB().Exec("INSERT INTO user_bank_accounts (user_id, account_type, account_name, account_no, bank_name, is_default, created_at) VALUES (?,?,?,?,?,?,NOW())",
		uid, req.Type, req.AccountName, req.AccountNo, req.BankName, req.IsDefault)
	if tx.Error != nil {
		response.Error(c, http.StatusBadRequest, "新增提现账户失败")
		return
	}
	response.Success(c, gin.H{"id": tx.RowsAffected})
}

// DeleteMyBankAccount DELETE /api/v1/wallet/bank-accounts/:id
func DeleteMyBankAccount(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未获取到用户身份")
		return
	}
	bid, _ := strconv.Atoi(c.Param("id"))
	if bid <= 0 {
		response.BadRequest(c, "非法ID")
		return
	}
	tx := database.GetDB().Exec("DELETE FROM user_bank_accounts WHERE id = ? AND user_id = ?", bid, uid)
	if tx.Error != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}
	response.Success(c, gin.H{"deleted": tx.RowsAffected})
}

// ===== 用户提现申请 =====

type withdrawalRow struct {
	ID            uint   `json:"id"`
	UserID        uint   `json:"user_id"`
	AmountCents   int64  `json:"amount_cents"`
	FeeCents      int64  `json:"fee_cents"`
	Status        string `json:"status"`
	BankAccountID *uint  `json:"bank_account_id"`
	Remark        string `json:"remark"`
	RequestedAt   string `json:"requested_at"`
	ProcessedAt   string `json:"processed_at,omitempty"`
}

// CreateMyWithdrawal POST /api/v1/wallet/withdrawals （当前用户）
func CreateMyWithdrawal(c *gin.Context) { createUserWithdrawalCommon(c, true) }

// CreateUserWithdrawal POST /api/v1/users/:id/withdrawals
func CreateUserWithdrawal(c *gin.Context) { createUserWithdrawalCommon(c, false) }

func createUserWithdrawalCommon(c *gin.Context, useCurrent bool) {
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未获取到用户身份")
		return
	}
	var pathUID uint = uid
	if !useCurrent {
		pid, _ := strconv.Atoi(c.Param("id"))
		if pid <= 0 || uint(pid) != uid {
			response.Forbidden(c, "不可为他人账户发起提现")
			return
		}
		pathUID = uint(pid)
	}
	var req struct {
		BankAccountID uint   `json:"bank_account_id"`
		AmountCents   int64  `json:"amount_cents"`
		Currency      string `json:"currency"`
		Note          string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.AmountCents <= 0 {
		response.BadRequest(c, "参数错误或金额需大于0")
		return
	}
	// 最低提现额度校验
	wf := config.Config.Finance.Withdrawal
	if wf.MinAmountCents > 0 && req.AmountCents < wf.MinAmountCents {
		response.BadRequest(c, "低于最低提现额度")
		return
	}
	// 校验余额（简化版）：读取 wallets.balance（分）
	type wrow struct{ Balance int64 }
	var w wrow
	if err := database.GetDB().Table("wallets").Select("balance").Where("user_id = ?", uid).Take(&w).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "钱包余额不可用")
		return
	}
	if w.Balance < req.AmountCents {
		response.BadRequest(c, "余额不足")
		return
	}
	// 校验提现账户归属（如传入）
	if req.BankAccountID != 0 {
		var cnt int64
		if err := database.GetDB().Table("user_bank_accounts").Where("id = ? AND user_id = ?", req.BankAccountID, pathUID).Count(&cnt).Error; err != nil || cnt == 0 {
			response.BadRequest(c, "提现账户无效或不属于当前用户")
			return
		}
	}

	// 计算手续费与净额
	fee := calcWithdrawalFee(req.AmountCents)
	if fee < 0 {
		fee = 0
	}
	if fee >= req.AmountCents {
		response.BadRequest(c, "手续费不合法，净额不得为0或负数")
		return
	}
	net := req.AmountCents - fee

	// 冻结资金：将可用余额转入冻结，减少并发风险
	// 读取当前可用余额
	var cur struct {
		Balance int64
		Frozen  int64
	}
	if err := database.GetDB().Table("wallets").Select("balance,frozen").Where("user_id = ?", pathUID).Take(&cur).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "读取钱包失败")
		return
	}
	if cur.Balance < req.AmountCents {
		response.BadRequest(c, "余额不足（可用余额）")
		return
	}
	newBal := cur.Balance - req.AmountCents
	newFrozen := cur.Frozen + req.AmountCents
	if err := database.GetDB().Exec("UPDATE wallets SET balance = ?, frozen = ? WHERE user_id = ?", newBal, newFrozen, pathUID).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "冻结资金失败")
		return
	}
	// 记录冻结流水（可用余额减少），remark 统一为 JSON
	remarkJSON := buildFreezeRemark(req.AmountCents, fee, net)
	if err := database.GetDB().Exec("INSERT INTO wallet_transactions (user_id, type, amount, balance_after, remark, created_at) VALUES (?,?,?,?,?,NOW())",
		pathUID, "withdraw_freeze", -req.AmountCents, newBal, remarkJSON).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "记录冻结流水失败")
		return
	}

	// 创建提现请求记录（记录金额与手续费）
	tx := database.GetDB().Exec("INSERT INTO withdrawal_requests (user_id, amount, fee, status, bank_account_id, remark, requested_at) VALUES (?,?,?,?,?,?,NOW())",
		pathUID, req.AmountCents, fee, "pending", req.BankAccountID, req.Note)
	if tx.Error != nil {
		response.Error(c, http.StatusInternalServerError, "创建提现请求失败")
		return
	}
	response.Success(c, gin.H{
		"status":           "pending",
		"amount_cents":     req.AmountCents,
		"fee_cents":        fee,
		"net_amount_cents": net,
	})
}

// ListUserWithdrawals GET /api/v1/users/:id/withdrawals
func ListUserWithdrawals(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未获取到用户身份")
		return
	}
	pid, _ := strconv.Atoi(c.Param("id"))
	if pid <= 0 || uint(pid) != uid {
		response.Forbidden(c, "不可查看他人提现记录")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	type raw struct {
		ID            uint
		UserID        uint
		Amount        int64
		Fee           int64
		Status        string
		BankAccountID *uint
		Remark        string
		RequestedAt   time.Time
		ProcessedAt   *time.Time
	}
	var total int64
	q := database.GetDB().Table("withdrawal_requests").Where("user_id = ?", uid)
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "统计失败")
		return
	}
	var raws []raw
	if err := q.Order("id desc").Limit(limit).Offset((page - 1) * limit).Find(&raws).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	items := make([]withdrawalRow, 0, len(raws))
	for _, r := range raws {
		var processed string
		if r.ProcessedAt != nil {
			processed = r.ProcessedAt.Format(time.RFC3339)
		}
		items = append(items, withdrawalRow{
			ID: r.ID, UserID: r.UserID, AmountCents: r.Amount, FeeCents: r.Fee, Status: r.Status,
			BankAccountID: r.BankAccountID, Remark: r.Remark,
			RequestedAt: r.RequestedAt.Format(time.RFC3339), ProcessedAt: processed,
		})
	}
	response.SuccessWithPagination(c, items, total, page, limit)
}

// （管理端审批逻辑已在 `internal/handler/withdraw_admin.go` 中实现，
//  此文件仅包含用户端提现相关处理函数。）

// ===== 工具函数 =====

func currentUserID(c *gin.Context) (uint, bool) {
	if v, ok := c.Get("user_id"); ok {
		switch t := v.(type) {
		case uint:
			return t, true
		case int:
			return uint(t), true
		case int64:
			return uint(t), true
		case string:
			if id, err := strconv.ParseUint(t, 10, 64); err == nil {
				return uint(id), true
			}
		}
	}
	return 0, false
}

func maskAccount(s string) string {
	if len(s) <= 4 {
		return s
	}
	tail := s[len(s)-4:]
	return "****" + tail
}

// calcWithdrawalFee 根据配置计算提现手续费（分）
// 手续费 = max(固定手续费, 按比例计算的手续费)，然后应用最低/封顶规则
func calcWithdrawalFee(amount int64) int64 {
	wf := config.Config.Finance.Withdrawal
	var rateFee int64 = 0
	if wf.FeeRateBp > 0 && amount > 0 {
		rateFee = (amount*wf.FeeRateBp + 9999) / 10000
	}
	fee := rateFee
	if wf.FeeFixedCents > fee {
		fee = wf.FeeFixedCents
	}
	if wf.FeeMinCents > 0 && fee < wf.FeeMinCents {
		fee = wf.FeeMinCents
	}
	if wf.FeeCapCents > 0 && fee > wf.FeeCapCents {
		fee = wf.FeeCapCents
	}
	if fee < 0 {
		fee = 0
	}
	return fee
}
