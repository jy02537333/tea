package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/response"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// CreateCategory 创建商品分类
func (h *ProductHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Sort        int    `json:"sort"`
		ParentID    uint   `json:"parent_id"`
		Image       string `json:"image"`
		Status      int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	category := &model.Category{
		Name:        req.Name,
		Description: req.Description,
		Sort:        req.Sort,
		ParentID:    req.ParentID,
		Image:       req.Image,
		Status:      req.Status,
	}

	if err := h.productService.CreateCategory(category); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建分类失败", err.Error())
		return
	}

	response.Success(c, category)
}

// GetCategories 获取商品分类列表
func (h *ProductHandler) GetCategories(c *gin.Context) {
	parentID := c.Query("parent_id")
	status := c.Query("status")

	var parentIDPtr *uint
	if parentID != "" {
		if id, err := strconv.ParseUint(parentID, 10, 32); err == nil {
			pid := uint(id)
			parentIDPtr = &pid
		}
	}

	var statusPtr *int
	if status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			statusPtr = &s
		}
	}

	categories, err := h.productService.GetCategories(parentIDPtr, statusPtr)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取分类失败", err.Error())
		return
	}

	response.Success(c, categories)
}

// UpdateCategory 更新商品分类
func (h *ProductHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的分类ID", err.Error())
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Sort        int    `json:"sort"`
		ParentID    uint   `json:"parent_id"`
		Image       string `json:"image"`
		Status      int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"sort":        req.Sort,
		"parent_id":   req.ParentID,
		"image":       req.Image,
		"status":      req.Status,
		"updated_at":  time.Now(),
	}

	if err := h.productService.UpdateCategory(uint(id), updates); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新分类失败", err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteCategory 删除商品分类
func (h *ProductHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的分类ID", err.Error())
		return
	}

	if err := h.productService.DeleteCategory(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除分类失败", err.Error())
		return
	}

	response.Success(c, nil)
}

// CreateProduct 创建商品
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	// 兼容 price/original_price 传入为数字或字符串
	var raw map[string]any
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	name, _ := raw["name"].(string)
	if name == "" {
		response.Error(c, http.StatusBadRequest, "名称必填")
		return
	}

	// category_id 可能是数字
	var categoryID uint
	switch v := raw["category_id"].(type) {
	case float64:
		if v < 0 {
			response.Error(c, http.StatusBadRequest, "非法的分类ID")
			return
		}
		categoryID = uint(v)
	case string:
		if v == "" {
			response.Error(c, http.StatusBadRequest, "分类必填")
			return
		}
		if id64, err := strconv.ParseUint(v, 10, 32); err == nil {
			categoryID = uint(id64)
		} else {
			response.Error(c, http.StatusBadRequest, "非法的分类ID")
			return
		}
	default:
		response.Error(c, http.StatusBadRequest, "分类必填")
		return
	}

	parseDecimal := func(val any) (decimal.Decimal, bool) {
		switch x := val.(type) {
		case float64:
			return decimal.NewFromFloat(x), true
		case string:
			d, err := decimal.NewFromString(x)
			if err != nil {
				return decimal.Zero, false
			}
			return d, true
		case nil:
			return decimal.Zero, true
		default:
			return decimal.Zero, false
		}
	}

	price, ok := parseDecimal(raw["price"])
	if !ok {
		response.Error(c, http.StatusBadRequest, "非法的价格格式")
		return
	}
	origPrice, ok := parseDecimal(raw["original_price"])
	if !ok {
		response.Error(c, http.StatusBadRequest, "非法的原价格式")
		return
	}

	// 其他字段
	description, _ := raw["description"].(string)
	images, _ := raw["images"].(string)
	status := 0
	if f, ok := raw["status"].(float64); ok {
		status = int(f)
	}
	stock := 0
	if f, ok := raw["stock"].(float64); ok {
		stock = int(f)
	}
	sales := 0
	if f, ok := raw["sales"].(float64); ok {
		sales = int(f)
	}
	sort := 0
	if f, ok := raw["sort"].(float64); ok {
		sort = int(f)
	}
	isHot := false
	if b, ok := raw["is_hot"].(bool); ok {
		isHot = b
	}
	isNew := false
	if b, ok := raw["is_new"].(bool); ok {
		isNew = b
	}
	isRecommend := false
	if b, ok := raw["is_recommend"].(bool); ok {
		isRecommend = b
	}

	product := &model.Product{
		Name:          name,
		Description:   description,
		CategoryID:    categoryID,
		Price:         price,
		OriginalPrice: origPrice,
		Images:        images,
		Status:        status,
		Stock:         stock,
		Sales:         sales,
		Sort:          sort,
		IsHot:         isHot,
		IsNew:         isNew,
		IsRecommend:   isRecommend,
	}

	if err := h.productService.CreateProduct(product); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建商品失败", err.Error())
		return
	}

	response.Success(c, product)
}

// GetProducts 获取商品列表
func (h *ProductHandler) GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	categoryID := c.Query("category_id")
	status := c.Query("status")
	keyword := c.Query("keyword")
	storeIDStr := c.Query("store_id")

	var categoryIDPtr *uint
	if categoryID != "" {
		if id, err := strconv.ParseUint(categoryID, 10, 32); err == nil {
			cid := uint(id)
			categoryIDPtr = &cid
		}
	}

	// 如传入 store_id，则返回包含门店库存与覆盖价的商品列表
	if storeIDStr != "" {
		if sid, err := strconv.ParseUint(storeIDStr, 10, 32); err == nil && sid > 0 {
			products, total, err := h.productService.GetProductsForStore(page, limit, categoryIDPtr, status, keyword, uint(sid))
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "获取商品失败", err.Error())
				return
			}
			response.SuccessWithPagination(c, products, total, page, limit)
			return
		}
	}

	// 默认返回普通商品列表
	products, total, err := h.productService.GetProducts(page, limit, categoryIDPtr, status, keyword)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取商品失败", err.Error())
		return
	}

	response.SuccessWithPagination(c, products, total, page, limit)
}

// GetProduct 获取商品详情
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的商品ID", err.Error())
		return
	}
	if sidStr := c.Query("store_id"); sidStr != "" {
		if sid, err2 := strconv.ParseUint(sidStr, 10, 32); err2 == nil && sid > 0 {
			product, err := h.productService.GetProductForStore(uint(id), uint(sid))
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "获取商品详情失败", err.Error())
				return
			}
			response.Success(c, product)
			return
		}
	}

	product, err := h.productService.GetProduct(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取商品详情失败", err.Error())
		return
	}

	response.Success(c, product)
}

// UpdateProduct 更新商品
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的商品ID", err.Error())
		return
	}

	var req struct {
		Name          string  `json:"name"`
		Description   string  `json:"description"`
		CategoryID    uint    `json:"category_id"`
		Price         float64 `json:"price"`
		OriginalPrice float64 `json:"original_price"`
		Images        string  `json:"images"`
		Status        int     `json:"status"`
		Stock         int     `json:"stock"`
		Sort          int     `json:"sort"`
		IsHot         bool    `json:"is_hot"`
		IsNew         bool    `json:"is_new"`
		IsRecommend   bool    `json:"is_recommend"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	updates := map[string]interface{}{
		"name":           req.Name,
		"description":    req.Description,
		"category_id":    req.CategoryID,
		"price":          decimal.NewFromFloat(req.Price),
		"original_price": decimal.NewFromFloat(req.OriginalPrice),
		"images":         req.Images,
		"status":         req.Status,
		"stock":          req.Stock,
		"sort":           req.Sort,
		"is_hot":         req.IsHot,
		"is_new":         req.IsNew,
		"is_recommend":   req.IsRecommend,
		"updated_at":     time.Now(),
	}

	if err := h.productService.UpdateProduct(uint(id), updates); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新商品失败", err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteProduct 删除商品
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的商品ID", err.Error())
		return
	}

	if err := h.productService.DeleteProduct(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除商品失败", err.Error())
		return
	}

	response.Success(c, nil)
}

// UpdateProductStock 更新商品库存
func (h *ProductHandler) UpdateProductStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的商品ID", err.Error())
		return
	}

	var req struct {
		Stock  int    `json:"stock" binding:"required"`
		Action string `json:"action" binding:"required"` // add, sub, set
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	if err := h.productService.UpdateProductStock(uint(id), req.Stock, req.Action); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新库存失败", err.Error())
		return
	}

	response.Success(c, nil)
}

// GetProductImages 获取商品图片列表
func (h *ProductHandler) GetProductImages(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的商品ID", err.Error())
		return
	}
	imgs, err := h.productService.GetProductImages(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取商品图片失败", err.Error())
		return
	}
	response.Success(c, imgs)
}

// AddProductImage 添加商品图片（管理员）
func (h *ProductHandler) AddProductImage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的商品ID", err.Error())
		return
	}
	var req struct {
		ImageURL string `json:"image_url" binding:"required"`
		Sort     int    `json:"sort"`
		IsMain   bool   `json:"is_main"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	img, err := h.productService.AddProductImage(uint(id), req.ImageURL, req.Sort, req.IsMain)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "添加商品图片失败", err.Error())
		return
	}
	response.Success(c, img)
}

// UpdateProductImage 更新图片（排序/主图）
func (h *ProductHandler) UpdateProductImage(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的商品ID", err.Error())
		return
	}
	imgID64, err := strconv.ParseUint(c.Param("image_id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的图片ID", err.Error())
		return
	}
	var req struct {
		Sort   *int  `json:"sort"`
		IsMain *bool `json:"is_main"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}
	if req.Sort == nil && req.IsMain == nil {
		response.Error(c, http.StatusBadRequest, "至少提供一个要更新的字段 (sort 或 is_main)")
		return
	}
	if err := h.productService.UpdateProductImage(uint(imgID64), req.Sort, req.IsMain); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新图片失败", err.Error())
		return
	}
	response.Success(c, nil)
}

// DeleteProductImage 删除商品图片
func (h *ProductHandler) DeleteProductImage(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的商品ID", err.Error())
		return
	}
	imgID64, err := strconv.ParseUint(c.Param("image_id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的图片ID", err.Error())
		return
	}
	if err := h.productService.DeleteProductImage(uint(imgID64)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除图片失败", err.Error())
		return
	}
	response.Success(c, nil)
}

// BatchDeleteProductImages 批量删除商品图片（管理员）
func (h *ProductHandler) BatchDeleteProductImages(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的商品ID", err.Error())
		return
	}

	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	if len(req.IDs) == 0 {
		response.Success(c, nil)
		return
	}

	if err := h.productService.DeleteProductImages(req.IDs); err != nil {
		response.Error(c, http.StatusInternalServerError, "批量删除图片失败", err.Error())
		return
	}
	response.Success(c, nil)
}
