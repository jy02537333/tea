package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/response"
)

// StoreExclusiveProductsHandler 门店特供（商家商城）商品
// 约定：biz_type=3 代表“特供/门店用品/其他”。
// - GET: 返回本门店已绑定的特供商品（含库存、覆盖价）
// - POST: 绑定（或更新）特供商品库存与覆盖价
//
// 注意：当前实现基于 products + store_products 映射关系，不支持“完全独立于 products 的 store-only 商品”。
// 如需支持独立条目，应新增 store_only_products 等模型。
type StoreExclusiveProductsHandler struct {
	productSvc   *service.ProductService
	inventorySvc *service.StoreInventoryService
}

func NewStoreExclusiveProductsHandler(productSvc *service.ProductService) *StoreExclusiveProductsHandler {
	if productSvc == nil {
		productSvc = service.NewProductService()
	}
	return &StoreExclusiveProductsHandler{productSvc: productSvc, inventorySvc: service.NewStoreInventoryService()}
}

func (h *StoreExclusiveProductsHandler) List(c *gin.Context) {
	sid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || sid == 0 {
		response.BadRequest(c, "非法门店ID")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	keyword := c.Query("keyword")
	var categoryID *uint
	if v := c.Query("category_id"); v != "" {
		cid64, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			response.BadRequest(c, "非法分类ID")
			return
		}
		cid := uint(cid64)
		if cid > 0 {
			categoryID = &cid
		}
	}

	list, total, err := h.productSvc.GetExclusiveProductsForStore(page, limit, keyword, categoryID, uint(sid))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessWithPagination(c, list, total, page, limit)
}

func (h *StoreExclusiveProductsHandler) Create(c *gin.Context) {
	sid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || sid == 0 {
		response.BadRequest(c, "非法门店ID")
		return
	}
	var req struct {
		ProductID     uint   `json:"product_id" binding:"required"`
		Stock         int    `json:"stock"`
		PriceOverride string `json:"price_override"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	sp, err := h.inventorySvc.Upsert(uint(sid), req.ProductID, req.Stock, req.PriceOverride, 3)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, sp)
}

// CreateNew 为本门店直接创建“自有商品”（特供），并上架到该门店
// POST /api/v1/stores/:id/exclusive-products/new
func (h *StoreExclusiveProductsHandler) CreateNew(c *gin.Context) {
	sid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || sid == 0 {
		response.BadRequest(c, "非法门店ID")
		return
	}
	var req struct {
		Name           string `json:"name" binding:"required"`
		CategoryID     uint   `json:"category_id" binding:"required"`
		Price          string `json:"price" binding:"required"` // 元，字符串格式如 "19.90"
		Description    string `json:"description"`
		Images         string `json:"images"`
		Stock          int    `json:"stock"`                     // 门店库存
		PriceOverride  string `json:"price_override"`           // 门店售价覆盖（可选）
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	// 构造基础商品
	p := model.Product{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Images:      req.Images,
		Price:       decimal.NewFromInt(0),
		Status:      1,
	}
	// 解析价格
	dec, err := decimal.NewFromString(req.Price)
	if err != nil {
		response.BadRequest(c, "价格格式错误")
		return
	}
	p.Price = dec

	sp, err := h.productSvc.CreateExclusiveProductForStore(uint(sid), p, req.Stock, req.PriceOverride)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, sp)
}
