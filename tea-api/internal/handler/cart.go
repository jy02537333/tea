package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"tea-api/internal/service"
	"tea-api/pkg/response"
)

type CartHandler struct {
	svc *service.CartService
}

func NewCartHandler() *CartHandler {
	return &CartHandler{svc: service.NewCartService()}
}

type addCartItemReq struct {
	ProductID uint  `json:"product_id" binding:"required"`
	SkuID     *uint `json:"sku_id"`
	Quantity  int   `json:"quantity" binding:"required"`
}

// AddItem 添加购物车条目
func (h *CartHandler) AddItem(c *gin.Context) {
	var req addCartItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))

	item, err := h.svc.AddItem(userID, req.ProductID, req.SkuID, req.Quantity)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, item)
}

// List 列出购物车条目
func (h *CartHandler) List(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))

	items, err := h.svc.ListItems(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, items)
}

type updateCartItemReq struct {
	Quantity int `json:"quantity" binding:"required"`
}

// UpdateQuantity 更新购物车条目数量（<=0 删除）
func (h *CartHandler) UpdateQuantity(c *gin.Context) {
	var req updateCartItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))

	var itemID uint
	if err := bindUintParam(c, "id", &itemID); err != nil {
		response.BadRequest(c, "非法的ID")
		return
	}

	if err := h.svc.UpdateItem(userID, itemID, req.Quantity); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// Remove 移除购物车条目
func (h *CartHandler) Remove(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))
	var itemID uint
	if err := bindUintParam(c, "id", &itemID); err != nil {
		response.BadRequest(c, "非法的ID")
		return
	}
	if err := h.svc.RemoveItem(userID, itemID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// Clear 清空购物车
func (h *CartHandler) Clear(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))
	if err := h.svc.Clear(userID); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// bindUintParam 简单的 uint 路径参数绑定
func bindUintParam(c *gin.Context, key string, v *uint) error {
	// 使用 ShouldBindUri 的简化版（避免引入额外结构体）
	// 转换错误统一返回
	idStr := c.Param(key)
	var id uint
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		return err
	}
	*v = id
	return nil
}
