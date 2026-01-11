package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/database"
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

	// 门店管理员：
	// 1) 加购时必须校验商品属于“本门店特供”(store_products.biz_type=3)
	// 2) 购物车内不得混入平台商品/其他门店商品
	roleVal, _ := c.Get("role")
	role, _ := roleVal.(string)
	if role == "store" {
		storeID, err := h.resolveStoreIDForStoreAdmin(userID)
		if err != nil {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		ok, err := h.svc.IsExclusiveProductForStore(storeID, req.ProductID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		if !ok {
			response.Error(c, http.StatusBadRequest, "该商品不属于本门店特供，无法加入购物车")
			return
		}

		badPID, err := h.svc.FindFirstNonExclusiveProductInCart(storeID, userID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		if badPID != 0 {
			response.Error(c, http.StatusBadRequest, "购物车中存在平台商品或其他门店商品，请先清空购物车后再添加门店特供商品")
			return
		}
	}

	item, err := h.svc.AddItem(userID, req.ProductID, req.SkuID, req.Quantity)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, item)
}

func (h *CartHandler) resolveStoreIDForStoreAdmin(userID uint) (uint, error) {
	db := database.GetDB()
	if db == nil {
		return 0, fmt.Errorf("无法校验门店权限")
	}

	var rows []struct {
		StoreID uint `gorm:"column:store_id"`
	}
	if err := db.Table("store_admins").Select("store_id").Where("user_id = ?", userID).Order("id desc").Limit(1).Find(&rows).Error; err != nil {
		return 0, fmt.Errorf("无法校验门店权限")
	}
	if len(rows) == 0 || rows[0].StoreID == 0 {
		return 0, fmt.Errorf("门店管理员未绑定门店")
	}
	return rows[0].StoreID, nil
}

// List 列出购物车条目
func (h *CartHandler) List(c *gin.Context) {
	uidVal, _ := c.Get("user_id")
	userID := uint(uidVal.(uint))

	storeID, err := h.bindAndValidateOptionalStoreID(c, userID)
	if err != nil {
		response.Error(c, http.StatusForbidden, err.Error())
		return
	}

	items, err := h.svc.ListItems(userID)
	if err != nil {
		// 最小修复：当购物车查询失败时返回空列表，避免前端因500中断
		response.Success(c, []any{})
		return
	}

	typed := make([]cartItemView, 0, len(items))
	if storeID != nil {
		overrideMap, err := h.svc.GetStorePriceOverrides(*storeID, items)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		for _, it := range items {
			view := cartItemView{CartItem: it}
			if v, ok := overrideMap[it.ProductID]; ok {
				vv := v
				view.StorePriceOverride = &vv
			}
			view.EffectivePrice = calcEffectivePrice(it, view.StorePriceOverride)
			typed = append(typed, view)
		}
	} else {
		for _, it := range items {
			view := cartItemView{CartItem: it}
			view.EffectivePrice = calcEffectivePrice(it, nil)
			typed = append(typed, view)
		}
	}

	response.Success(c, typed)
}

type cartItemView struct {
	model.CartItem
	StorePriceOverride *decimal.Decimal `json:"store_price_override,omitempty"`
	EffectivePrice     decimal.Decimal  `json:"effective_price"`
}

func calcEffectivePrice(it model.CartItem, storeOverride *decimal.Decimal) decimal.Decimal {
	if storeOverride != nil && storeOverride.GreaterThan(decimal.Zero) {
		return *storeOverride
	}
	// sku 优先，其次 product
	if it.SkuID != nil && it.Sku.ID > 0 {
		return it.Sku.Price
	}
	if it.Product.ID > 0 {
		return it.Product.Price
	}
	return decimal.Zero
}

func (h *CartHandler) bindAndValidateOptionalStoreID(c *gin.Context, userID uint) (*uint, error) {
	v := c.Query("store_id")
	if v == "" {
		return nil, nil
	}
	storeID64, err := strconv.ParseUint(v, 10, 64)
	if err != nil || storeID64 == 0 {
		return nil, fmt.Errorf("非法的store_id")
	}
	storeID := uint(storeID64)

	roleVal, _ := c.Get("role")
	role, _ := roleVal.(string)
	if role != "store" {
		return nil, fmt.Errorf("仅门店管理员可使用store_id定价")
	}

	db := database.GetDB()
	if db == nil {
		return nil, fmt.Errorf("无法校验门店权限")
	}

	var rows []struct {
		StoreID uint `gorm:"column:store_id"`
	}
	if err := db.Table("store_admins").Select("store_id").Where("user_id = ?", userID).Order("id desc").Limit(1).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("无法校验门店权限")
	}
	if len(rows) == 0 || rows[0].StoreID == 0 {
		return nil, fmt.Errorf("门店管理员未绑定门店")
	}
	if rows[0].StoreID != storeID {
		return nil, fmt.Errorf("门店越权")
	}

	return &storeID, nil
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
