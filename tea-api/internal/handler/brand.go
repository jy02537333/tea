package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"tea-api/internal/model"
	"tea-api/internal/service"
	"tea-api/pkg/response"
)

type BrandHandler struct{ svc *service.BrandService }

func NewBrandHandler() *BrandHandler { return &BrandHandler{svc: service.NewBrandService()} }

// List 列出品牌
func (h *BrandHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    q := c.Query("q")
    list, total, err := h.svc.GetBrands(page, limit, q)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "获取品牌失败", err.Error())
        return
    }
    response.SuccessWithPagination(c, list, total, page, limit)
}

// Get 品牌详情
func (h *BrandHandler) Get(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil || id == 0 {
        response.Error(c, http.StatusBadRequest, "非法的品牌ID")
        return
    }
    b, err := h.svc.GetBrand(uint(id))
    if err != nil {
        response.Error(c, http.StatusNotFound, "品牌不存在", err.Error())
        return
    }
    response.Success(c, b)
}

// Create 创建品牌
func (h *BrandHandler) Create(c *gin.Context) {
    var req struct {
        Name           string `json:"name" binding:"required"`
        LogoURL        string `json:"logo_url"`
        OriginRegionID *uint  `json:"origin_region_id"`
        Description    string `json:"description"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
        return
    }
    b := &model.Brand{Name: req.Name, LogoURL: req.LogoURL, OriginRegionID: req.OriginRegionID, Description: req.Description}
    if err := h.svc.CreateBrand(b); err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }
    response.Success(c, b)
}

// Update 更新品牌
func (h *BrandHandler) Update(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil || id == 0 {
        response.Error(c, http.StatusBadRequest, "非法的品牌ID")
        return
    }
    var req struct {
        Name           *string `json:"name"`
        LogoURL        *string `json:"logo_url"`
        OriginRegionID *uint   `json:"origin_region_id"`
        Description    *string `json:"description"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "参数错误", err.Error())
        return
    }
    updates := map[string]any{"updated_at": time.Now()}
    if req.Name != nil {
        updates["name"] = *req.Name
    }
    if req.LogoURL != nil {
        updates["logo_url"] = *req.LogoURL
    }
    if req.OriginRegionID != nil {
        updates["origin_region_id"] = *req.OriginRegionID
    }
    if req.Description != nil {
        updates["description"] = *req.Description
    }
    if err := h.svc.UpdateBrand(uint(id), updates); err != nil {
        response.Error(c, http.StatusInternalServerError, "更新品牌失败", err.Error())
        return
    }
    response.Success(c, gin.H{"ok": true})
}

// Delete 删除品牌
func (h *BrandHandler) Delete(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil || id == 0 {
        response.Error(c, http.StatusBadRequest, "非法的品牌ID")
        return
    }
    if err := h.svc.DeleteBrand(uint(id)); err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }
    response.Success(c, gin.H{"ok": true})
}
