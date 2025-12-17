package handler

import (
	"github.com/gin-gonic/gin"

	"tea-api/internal/service/commission"
	"tea-api/pkg/utils"
)

type CommissionAdminHandler struct{}

func NewCommissionAdminHandler() *CommissionAdminHandler { return &CommissionAdminHandler{} }

type commissionReleaseReq struct {
	BatchSize *int `json:"batch_size"`
}

type commissionReverseOrderReq struct {
	OrderID uint `json:"order_id" binding:"required"`
}

// POST /api/v1/admin/commission/release
// 手动触发一次佣金解冻，仅限运维/财务使用
// 可选 body: {"batch_size": 100} 覆盖默认批次大小
func (h *CommissionAdminHandler) TriggerRelease(c *gin.Context) {
	var req commissionReleaseReq
	_ = c.ShouldBindJSON(&req)

	batch := 0
	if req.BatchSize != nil {
		batch = *req.BatchSize
	}

	processed, err := commission.ReleaseFrozenCommissions(batch)
	if err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	utils.Success(c, gin.H{"processed": processed})
}

// POST /api/v1/admin/finance/commission/reverse-order
// 根据订单ID 一键回滚该订单对应的未提现佣金（frozen/available -> reversed）
func (h *CommissionAdminHandler) ReverseOrder(c *gin.Context) {
	var req commissionReverseOrderReq
	if err := c.ShouldBindJSON(&req); err != nil || req.OrderID == 0 {
		utils.Error(c, utils.CodeError, "order_id 必填且必须大于0")
		return
	}

	var opIDPtr *uint
	if v, ok := c.Get("user_id"); ok {
		if u, ok2 := v.(uint); ok2 && u != 0 {
			opIDPtr = &u
		}
	}

	note := "manual commission rollback by admin"
	processed, err := commission.ReverseOrderCommissions(req.OrderID, opIDPtr, note)
	if err != nil {
		utils.Error(c, utils.CodeError, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"order_id":  req.OrderID,
		"processed": processed,
	})
}
