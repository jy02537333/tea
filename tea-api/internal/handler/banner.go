package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/response"
)

type BannerHandler struct{}

func NewBannerHandler() *BannerHandler { return &BannerHandler{} }

type bannerPayload struct {
	Title    string `json:"title"`
	ImageURL string `json:"image_url"`
	LinkType int    `json:"link_type"`
	LinkURL  string `json:"link_url"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status"`
}

// AdminListBanners 广告/轮播图列表（支持 keyword + status 查询）
func (h *BannerHandler) AdminListBanners(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	status := atoi(c.Query("status"))

	db := database.GetDB()
	q := db.Model(&model.Banner{})
	if keyword != "" {
		q = q.Where("title LIKE ?", "%"+keyword+"%")
	}
	if status == 1 || status == 2 {
		q = q.Where("status = ?", status)
	}

	var list []model.Banner
	if err := q.Order("sort desc").Order("id desc").Find(&list).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, list)
}

// AdminCreateBanner 新增广告/轮播图
func (h *BannerHandler) AdminCreateBanner(c *gin.Context) {
	var req bannerPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.ImageURL = strings.TrimSpace(req.ImageURL)
	req.LinkURL = strings.TrimSpace(req.LinkURL)
	if req.ImageURL == "" {
		response.BadRequest(c, "image_url 不能为空")
		return
	}
	if req.LinkType == 0 {
		req.LinkType = 1
	}
	if req.Status == 0 {
		req.Status = 1
	}

	b := model.Banner{
		Title:    req.Title,
		ImageURL: req.ImageURL,
		LinkType: req.LinkType,
		LinkURL:  req.LinkURL,
		Sort:     req.Sort,
		Status:   req.Status,
	}

	if err := database.GetDB().Create(&b).Error; err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, b)
}

// AdminUpdateBanner 编辑广告/轮播图（含上下架）
func (h *BannerHandler) AdminUpdateBanner(c *gin.Context) {
	id := uint(atoi(c.Param("id")))
	if id == 0 {
		response.BadRequest(c, "无效的ID")
		return
	}

	var req bannerPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updates := map[string]any{}
	if req.Title != "" {
		updates["title"] = strings.TrimSpace(req.Title)
	}
	if req.ImageURL != "" {
		updates["image_url"] = strings.TrimSpace(req.ImageURL)
	}
	if req.LinkType != 0 {
		updates["link_type"] = req.LinkType
	}
	if req.LinkURL != "" {
		updates["link_url"] = strings.TrimSpace(req.LinkURL)
	}
	updates["sort"] = req.Sort
	if req.Status == 1 || req.Status == 2 {
		updates["status"] = req.Status
	}

	if err := database.GetDB().Model(&model.Banner{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var b model.Banner
	if err := database.GetDB().First(&b, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, b)
}

// AdminDeleteBanner 删除广告/轮播图
func (h *BannerHandler) AdminDeleteBanner(c *gin.Context) {
	id := uint(atoi(c.Param("id")))
	if id == 0 {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := database.GetDB().Delete(&model.Banner{}, id).Error; err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}
