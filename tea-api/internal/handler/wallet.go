package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"tea-api/pkg/database"
	"tea-api/pkg/response"
)

// MyWalletSummary 用户钱包概要响应
// 金额单位统一为分（整数），与 PRD/OpenAPI 约定保持一致。
type MyWalletSummary struct {
	BalanceCents int64  `json:"balance_cents"`
	FrozenCents  int64  `json:"frozen_cents"`
	Currency     string `json:"currency"` // 默认 CNY
}

// GetMyWallet GET /api/v1/wallet
// 返回当前登录用户的钱包余额/冻结金额（分）。
func GetMyWallet(c *gin.Context) {
	uidVal, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未获取到用户身份")
		return
	}
	var userID uint
	switch v := uidVal.(type) {
	case uint:
		userID = v
	case int:
		userID = uint(v)
	case int64:
		userID = uint(v)
	case string:
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			userID = uint(id)
		}
	}
	if userID == 0 {
		response.Unauthorized(c, "未获取到用户身份")
		return
	}

	type walletRow struct {
		Balance int64
		Frozen  int64
	}
	var row walletRow
	if err := database.GetDB().Table("wallets").Select("balance, frozen").Where("user_id = ?", userID).Take(&row).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "查询钱包失败")
		return
	}
	response.Success(c, MyWalletSummary{BalanceCents: row.Balance, FrozenCents: row.Frozen, Currency: "CNY"})
}

// WalletTransaction 钱包流水条目（单位：分）
type WalletTransaction struct {
	ID           uint   `json:"id"`
	Type         string `json:"type"`
	Amount       int64  `json:"amount_cents"`
	BalanceAfter int64  `json:"balance_after_cents"`
	Remark       string `json:"remark"`
	CreatedAt    string `json:"created_at"`
}

// ListMyWalletTransactions GET /api/v1/wallet/transactions?page=&limit=
// 返回当前用户的钱包流水列表（分页）。
func ListMyWalletTransactions(c *gin.Context) {
	uidVal, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "未获取到用户身份")
		return
	}
	var userID uint
	switch v := uidVal.(type) {
	case uint:
		userID = v
	case int:
		userID = uint(v)
	case int64:
		userID = uint(v)
	case string:
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			userID = uint(id)
		}
	}
	if userID == 0 {
		response.Unauthorized(c, "未获取到用户身份")
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

	type txRow struct {
		ID           uint
		Type         string
		Amount       int64
		BalanceAfter *int64
		Remark       string
		CreatedAt    string
	}
	var total int64
	q := database.GetDB().Table("wallet_transactions").Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "统计流水数量失败")
		return
	}

	var rows []txRow
	if err := q.Order("id desc").Limit(limit).Offset((page - 1) * limit).Find(&rows).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "查询流水失败")
		return
	}

	items := make([]WalletTransaction, 0, len(rows))
	for _, r := range rows {
		var bal int64
		if r.BalanceAfter != nil {
			bal = *r.BalanceAfter
		}
		items = append(items, WalletTransaction{
			ID:           r.ID,
			Type:         r.Type,
			Amount:       r.Amount,
			BalanceAfter: bal,
			Remark:       r.Remark,
			CreatedAt:    r.CreatedAt,
		})
	}

	response.SuccessWithPagination(c, items, total, page, limit)
}
