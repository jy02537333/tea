package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"tea-api/internal/service"
	"tea-api/pkg/response"
)

type StoreInventoryHandler struct {
	svc *service.StoreInventoryService
}

func NewStoreInventoryHandler() *StoreInventoryHandler {
	return &StoreInventoryHandler{svc: service.NewStoreInventoryService()}
}

// Upsert 绑定或更新门店商品库存与价格
func (h *StoreInventoryHandler) Upsert(c *gin.Context) {
	sid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法门店ID")
		return
	}
	var req struct {
		ProductID     uint   `json:"product_id" binding:"required"`
		Stock         int    `json:"stock"`
		BizType       *int   `json:"biz_type"`       // 1:服务 2:外卖 3:其他；为空时默认 1
		PriceOverride string `json:"price_override"` // 可为空字符串表示无覆盖
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	bizType := 1
	if req.BizType != nil && *req.BizType > 0 {
		bizType = *req.BizType
	}
	sp, err := h.svc.Upsert(uint(sid), req.ProductID, req.Stock, req.PriceOverride, bizType)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, sp)
}

func (h *StoreInventoryHandler) List(c *gin.Context) {
	sid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法门店ID")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	bizTypeParam := c.Query("biz_type")
	var bizTypePtr *int
	if bizTypeParam != "" {
		if v, err := strconv.Atoi(bizTypeParam); err == nil && v > 0 {
			bizTypePtr = &v
		}
	}
	list, total, err := h.svc.List(uint(sid), page, limit, bizTypePtr)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessWithPagination(c, list, total, page, limit)
}

func (h *StoreInventoryHandler) Delete(c *gin.Context) {
	sid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法门店ID")
		return
	}
	pid, err := strconv.ParseUint(c.Param("pid"), 10, 32)
	if err != nil {
		response.BadRequest(c, "非法商品ID")
		return
	}
	if err := h.svc.Delete(uint(sid), uint(pid)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}
