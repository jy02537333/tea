package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tea-api/pkg/database"
	"tea-api/pkg/response"
)

type RechargeAdminHandler struct {
	db *gorm.DB
}

func NewRechargeAdminHandler() *RechargeAdminHandler {
	return &RechargeAdminHandler{db: database.GetDB()}
}

type adminWalletSummary struct {
	UserID       uint  `json:"user_id"`
	BalanceCents int64 `json:"balance_cents"`
	FrozenCents  int64 `json:"frozen_cents"`
}

type adminWalletTx struct {
	ID              uint   `json:"id"`
	UserID          uint   `json:"user_id"`
	Type            string `json:"type"`
	AmountCents     int64  `json:"amount_cents"`
	BalanceAfterCts int64  `json:"balance_after_cents"`
	Remark          string `json:"remark"`
	CreatedAt       string `json:"created_at"`
}

type walletActionReq struct {
	AmountCents int64  `json:"amount_cents" binding:"required"`
	Remark      string `json:"remark"`
}

func parseUintParam(c *gin.Context, name string) uint {
	v := strings.TrimSpace(c.Param(name))
	if v == "" {
		v = strings.TrimSpace(c.Query(name))
	}
	if v == "" {
		return 0
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0
	}
	return uint(n)
}

func (h *RechargeAdminHandler) ensureWalletRow(tx *gorm.DB, userID uint) error {
	// ensure user exists
	var cnt int64
	if err := tx.Table("users").Where("id = ?", userID).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return gorm.ErrRecordNotFound
	}

	// wallets is keyed by user_id; insert if missing
	return tx.Exec("INSERT IGNORE INTO wallets (user_id, balance, frozen) VALUES (?,?,?)", userID, 0, 0).Error
}

func (h *RechargeAdminHandler) getWallet(tx *gorm.DB, userID uint) (balance, frozen int64, err error) {
	type walletRow struct {
		Balance int64
		Frozen  int64
	}
	var row walletRow
	if err := h.ensureWalletRow(tx, userID); err != nil {
		return 0, 0, err
	}
	if err := tx.Table("wallets").Select("balance,frozen").Where("user_id = ?", userID).Take(&row).Error; err != nil {
		return 0, 0, err
	}
	return row.Balance, row.Frozen, nil
}

// ListRecords GET /api/v1/admin/recharge/records?user_id=&types=a,b&keyword=&start=&end=&page=&limit=
func (h *RechargeAdminHandler) ListRecords(c *gin.Context) {
	if h.db == nil {
		response.Error(c, http.StatusInternalServerError, "db not ready")
		return
	}

	userID := parseUintParam(c, "user_id")
	typesParam := strings.TrimSpace(c.Query("types"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))

	page := toInt(c.DefaultQuery("page", "1"))
	limit := toInt(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}

	q := h.db.Table("wallet_transactions")
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if typesParam != "" {
		raw := strings.Split(typesParam, ",")
		types := make([]string, 0, len(raw))
		for _, t := range raw {
			t = strings.TrimSpace(t)
			if t != "" {
				types = append(types, t)
			}
		}
		if len(types) > 0 {
			q = q.Where("type IN ?", types)
		}
	}
	if keyword != "" {
		q = q.Where("remark LIKE ?", "%"+keyword+"%")
	}
	if start != "" {
		q = q.Where("created_at >= ?", start)
	}
	if end != "" {
		q = q.Where("created_at <= ?", end)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	type txRow struct {
		ID           uint
		UserID       uint
		Type         string
		Amount       int64
		BalanceAfter *int64
		Remark       string
		CreatedAt    string
	}
	var rows []txRow
	if err := q.Order("id desc").Limit(limit).Offset((page - 1) * limit).Find(&rows).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	items := make([]adminWalletTx, 0, len(rows))
	for _, r := range rows {
		var bal int64
		if r.BalanceAfter != nil {
			bal = *r.BalanceAfter
		}
		items = append(items, adminWalletTx{
			ID:              r.ID,
			UserID:          r.UserID,
			Type:            r.Type,
			AmountCents:     r.Amount,
			BalanceAfterCts: bal,
			Remark:          r.Remark,
			CreatedAt:       r.CreatedAt,
		})
	}

	response.SuccessWithPagination(c, items, total, page, limit)
}

// GetUserWallet GET /api/v1/admin/recharge/users/:id/wallet
func (h *RechargeAdminHandler) GetUserWallet(c *gin.Context) {
	if h.db == nil {
		response.Error(c, http.StatusInternalServerError, "db not ready")
		return
	}
	userID := parseUintParam(c, "id")
	if userID == 0 {
		response.BadRequest(c, "参数错误")
		return
	}

	balance, frozen, err := h.getWallet(h.db, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.BadRequest(c, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"wallet": adminWalletSummary{UserID: userID, BalanceCents: balance, FrozenCents: frozen},
	})
}

func (h *RechargeAdminHandler) mutateWallet(c *gin.Context, typ string, deltaBalance int64, deltaFrozen int64, remark string) {
	if h.db == nil {
		response.Error(c, http.StatusInternalServerError, "db not ready")
		return
	}
	userID := parseUintParam(c, "id")
	if userID == 0 {
		response.BadRequest(c, "参数错误")
		return
	}
	remark = strings.TrimSpace(remark)
	if remark == "" {
		remark = typ
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		bal, fro, err := h.getWallet(tx, userID)
		if err != nil {
			return err
		}
		newBal := bal + deltaBalance
		newFrozen := fro + deltaFrozen
		if newBal < 0 || newFrozen < 0 {
			return gorm.ErrInvalidData
		}
		if err := tx.Exec("UPDATE wallets SET balance = ?, frozen = ? WHERE user_id = ?", newBal, newFrozen, userID).Error; err != nil {
			return err
		}
		// wallet_transactions.amount records the balance delta
		if err := tx.Exec(
			"INSERT INTO wallet_transactions (user_id, type, amount, balance_after, remark, created_at) VALUES (?,?,?,?,?,NOW())",
			userID, typ, deltaBalance, newBal, remark,
		).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.BadRequest(c, "用户不存在")
			return
		}
		if err == gorm.ErrInvalidData {
			response.BadRequest(c, "余额不足")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	balance, frozen, err2 := h.getWallet(h.db, userID)
	if err2 != nil {
		response.Error(c, http.StatusInternalServerError, err2.Error())
		return
	}
	response.Success(c, gin.H{
		"ok":     true,
		"wallet": adminWalletSummary{UserID: userID, BalanceCents: balance, FrozenCents: frozen},
	})
}

// Freeze POST /api/v1/admin/recharge/users/:id/freeze
func (h *RechargeAdminHandler) Freeze(c *gin.Context) {
	var req walletActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.AmountCents <= 0 {
		response.BadRequest(c, "amount_cents 必须大于 0")
		return
	}
	h.mutateWallet(c, "freeze", -req.AmountCents, +req.AmountCents, req.Remark)
}

// Unfreeze POST /api/v1/admin/recharge/users/:id/unfreeze
func (h *RechargeAdminHandler) Unfreeze(c *gin.Context) {
	var req walletActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.AmountCents <= 0 {
		response.BadRequest(c, "amount_cents 必须大于 0")
		return
	}
	h.mutateWallet(c, "unfreeze", +req.AmountCents, -req.AmountCents, req.Remark)
}

// Credit POST /api/v1/admin/recharge/users/:id/credit
func (h *RechargeAdminHandler) Credit(c *gin.Context) {
	var req walletActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.AmountCents <= 0 {
		response.BadRequest(c, "amount_cents 必须大于 0")
		return
	}
	h.mutateWallet(c, "admin_credit", +req.AmountCents, 0, req.Remark)
}

// Debit POST /api/v1/admin/recharge/users/:id/debit
func (h *RechargeAdminHandler) Debit(c *gin.Context) {
	var req walletActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.AmountCents <= 0 {
		response.BadRequest(c, "amount_cents 必须大于 0")
		return
	}
	h.mutateWallet(c, "admin_debit", -req.AmountCents, 0, req.Remark)
}
