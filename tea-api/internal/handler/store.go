package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/response"
)

type StoreHandler struct{ svc *service.StoreService }

func NewStoreHandler() *StoreHandler { return &StoreHandler{svc: service.NewStoreService()} }

// ListAccounts 列出门店收款账户
// GET /api/v1/stores/:id/accounts
func (h *StoreHandler) ListAccounts(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "非法ID")
		return
	}

	list, err := h.svc.ListStoreAccounts(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, list)
}

// CreateAccount 为门店新增收款账户
// POST /api/v1/stores/:id/accounts
func (h *StoreHandler) CreateAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "非法ID")
		return
	}

	var req struct {
		AccountType string `json:"account_type"`
		AccountName string `json:"account_name" binding:"required"`
		AccountNo   string `json:"account_no" binding:"required"`
		BankName    string `json:"bank_name"`
		IsDefault   bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	acc, err := h.svc.CreateStoreAccount(uint(id), req.AccountType, req.AccountName, req.AccountNo, req.BankName, req.IsDefault)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, acc)
}

// UpdateAccount 更新门店收款账户
// PUT /api/v1/stores/:id/accounts/:accountId
func (h *StoreHandler) UpdateAccount(c *gin.Context) {
	storeID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || storeID == 0 {
		response.BadRequest(c, "非法ID")
		return
	}
	accID, err := strconv.ParseUint(c.Param("accountId"), 10, 32)
	if err != nil || accID == 0 {
		response.BadRequest(c, "非法账户ID")
		return
	}

	var req struct {
		AccountType *string `json:"account_type"`
		AccountName *string `json:"account_name"`
		AccountNo   *string `json:"account_no"`
		BankName    *string `json:"bank_name"`
		IsDefault   *bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.svc.UpdateStoreAccount(uint(storeID), uint(accID), req.AccountType, req.AccountName, req.AccountNo, req.BankName, req.IsDefault); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// DeleteAccount 删除门店收款账户
// DELETE /api/v1/stores/:id/accounts/:accountId
func (h *StoreHandler) DeleteAccount(c *gin.Context) {
	storeID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || storeID == 0 {
		response.BadRequest(c, "非法ID")
		return
	}
	accID, err := strconv.ParseUint(c.Param("accountId"), 10, 32)
	if err != nil || accID == 0 {
		response.BadRequest(c, "非法账户ID")
		return
	}

	if err := h.svc.DeleteStoreAccount(uint(storeID), uint(accID)); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

func (h *StoreHandler) Create(c *gin.Context) {
	var req struct {
		Name          string  `json:"name" binding:"required"`
		Address       string  `json:"address"`
		Phone         string  `json:"phone"`
		Latitude      float64 `json:"latitude"`
		Longitude     float64 `json:"longitude"`
		BusinessHours string  `json:"business_hours"`
		Images        string  `json:"images"`
		Status        int     `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	st := &model.Store{Name: req.Name, Address: req.Address, Phone: req.Phone, Latitude: req.Latitude, Longitude: req.Longitude, BusinessHours: req.BusinessHours, Images: req.Images, Status: req.Status}
	if err := h.svc.CreateStore(st); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, st)
}

func (h *StoreHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法ID")
		return
	}
	var req struct {
		Name          string  `json:"name"`
		Address       string  `json:"address"`
		Phone         string  `json:"phone"`
		Latitude      float64 `json:"latitude"`
		Longitude     float64 `json:"longitude"`
		BusinessHours string  `json:"business_hours"`
		Images        string  `json:"images"`
		Status        int     `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	updates := map[string]any{
		"name":           req.Name,
		"address":        req.Address,
		"phone":          req.Phone,
		"latitude":       req.Latitude,
		"longitude":      req.Longitude,
		"business_hours": req.BusinessHours,
		"images":         req.Images,
		"status":         req.Status,
		"updated_at":     time.Now(),
	}
	if err := h.svc.UpdateStore(uint(id), updates); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

func (h *StoreHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法ID")
		return
	}
	if err := h.svc.DeleteStore(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

func (h *StoreHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法ID")
		return
	}
	st, err := h.svc.GetStore(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	response.Success(c, st)
}

func (h *StoreHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	var statusPtr *int
	if v := c.Query("status"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			statusPtr = &n
		}
	}
	var latPtr, lngPtr *float64
	if v := c.Query("lat"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			latPtr = &f
		}
	}
	if v := c.Query("lng"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			lngPtr = &f
		}
	}

	list, total, err := h.svc.ListStores(page, limit, statusPtr, latPtr, lngPtr)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessWithPagination(c, list, total, page, limit)
}

// OrderStats 门店订单统计（管理员）
func (h *StoreHandler) OrderStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法ID")
		return
	}
	ordSvc := service.NewOrderService()
	stats, err := ordSvc.GetStoreOrderStats(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, stats)
}

// FinanceTransactions 门店资金流水列表
// GET /api/v1/stores/:id/finance/transactions?start=&end=&page=&limit=&type=
func (h *StoreHandler) FinanceTransactions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "非法ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))
	txType := strings.TrimSpace(c.Query("type")) // 可选：payment/refund/withdraw

	list, total, err := h.svc.ListStoreFinanceTransactions(uint(id), start, end, page, limit, txType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessWithPagination(c, list, total, page, limit)
}

// ExportFinanceTransactions 门店资金流水导出（CSV 雏形）
// GET /api/v1/stores/:id/finance/transactions/export?start=&end=&type=
func (h *StoreHandler) ExportFinanceTransactions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "非法ID")
		return
	}

	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))
	txType := strings.TrimSpace(c.Query("type"))

	// 导出不分页，简单按当前条件取全部，若未来有需要可限制时间跨度
	list, _, err := h.svc.ListStoreFinanceTransactions(uint(id), start, end, 1, 10000, txType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	var buf bytes.Buffer
	// 表头
	buf.WriteString("id,type,direction,amount,fee,method,related_id,related_no,remark,created_at\n")
	for _, item := range list {
		line := fmt.Sprintf("%d,%s,%s,%.2f,%.2f,%d,%d,%s,%s,%s\n",
			item.ID,
			item.Type,
			item.Direction,
			item.Amount.InexactFloat64(),
			item.Fee.InexactFloat64(),
			item.Method,
			item.RelatedID,
			escapeCSV(item.RelatedNo),
			escapeCSV(item.Remark),
			item.CreatedAt.Format("2006-01-02 15:04:05"),
		)
		buf.WriteString(line)
	}

	filename := fmt.Sprintf("store_%d_finance_%s.csv", id, time.Now().Format("20060102150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.String(http.StatusOK, buf.String())
}

// escapeCSV 简单处理逗号和换行符
func escapeCSV(s string) string {
	if strings.ContainsAny(s, ",\n\r\"") {
		// 用双引号包裹并转义内部双引号
		s = strings.ReplaceAll(s, "\"", "\"\"")
		return "\"" + s + "\""
	}
	return s
}

// Wallet 门店钱包概要（预计可提现余额）
// GET /api/v1/stores/:id/wallet
func (h *StoreHandler) Wallet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "非法ID")
		return
	}

	summary, err := h.svc.GetStoreWalletSummary(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, summary)
}

// ListWithdraws 门店提现记录列表
// GET /api/v1/stores/:id/withdraws?page=&limit=&status=
func (h *StoreHandler) ListWithdraws(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "非法ID")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	statusStr := strings.TrimSpace(c.Query("status"))
	var status *int
	if statusStr != "" {
		if v, errConv := strconv.Atoi(statusStr); errConv == nil {
			status = &v
		}
	}

	list, total, err := h.svc.ListStoreWithdraws(uint(id), page, limit, status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessWithPagination(c, list, total, page, limit)
}

// ApplyWithdraw 门店发起提现申请
// POST /api/v1/stores/:id/withdraws
// body: { "amount": 100.00, "remark": "备注", "withdraw_type": 1 }
func (h *StoreHandler) ApplyWithdraw(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "非法ID")
		return
	}

	var req struct {
		Amount       float64 `json:"amount" binding:"required"`
		Remark       string  `json:"remark"`
		WithdrawType int     `json:"withdraw_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.WithdrawType == 0 {
		req.WithdrawType = 1
	}
	if req.Amount <= 0 {
		response.BadRequest(c, "提现金额必须大于0")
		return
	}

	amount := decimal.NewFromFloat(req.Amount)
	var operatorID uint
	if v, ok := c.Get("user_id"); ok {
		if u, ok2 := v.(uint); ok2 {
			operatorID = u
		}
	}

	rec, err := h.svc.ApplyStoreWithdraw(uint(id), amount, req.Remark, req.WithdrawType, operatorID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, rec)
}
