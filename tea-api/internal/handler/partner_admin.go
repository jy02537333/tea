package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/response"
)

// PartnerAdminHandler 管理端合伙人管理（最小闭环）
type PartnerAdminHandler struct{}

func NewPartnerAdminHandler() *PartnerAdminHandler { return &PartnerAdminHandler{} }

// ListPartners GET /api/v1/admin/partners
func (h *PartnerAdminHandler) ListPartners(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    if page < 1 {
        page = 1
    }
    if limit <= 0 || limit > 200 {
        limit = 20
    }

    q := strings.TrimSpace(c.Query("q"))
    level := strings.TrimSpace(c.Query("level"))

    db := database.GetDB()
    base := db.Table("users").Select("users.id, users.nickname, users.phone, users.partner_level_id, users.created_at").Where("partner_level_id IS NOT NULL")
    if q != "" {
        like := "%" + q + "%"
        base = base.Where("nickname LIKE ? OR phone LIKE ?", like, like)
    }
    if level != "" {
        if l, err := strconv.Atoi(level); err == nil && l > 0 {
            base = base.Where("partner_level_id = ?", l)
        }
    }

    var total int64
    if err := base.Count(&total).Error; err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    type partnerRow struct {
        ID             uint   `json:"id"`
        Nickname       string `json:"nickname"`
        Phone          string `json:"phone"`
        PartnerLevelID *uint  `json:"partner_level_id"`
        CreatedAt      string `json:"created_at"`
    }

    var list []partnerRow
    if err := base.Order("id desc").Limit(limit).Offset((page-1)*limit).Scan(&list).Error; err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    response.SuccessWithPagination(c, list, total, page, limit)
}

// ListCommissions GET /api/v1/admin/partners/:id/commissions
func (h *PartnerAdminHandler) ListCommissions(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil || id == 0 {
        response.BadRequest(c, "非法ID")
        return
    }
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    if page < 1 {
        page = 1
    }
    if limit <= 0 || limit > 200 {
        limit = 20
    }

    db := database.GetDB()
    q := db.Model(&model.Commission{}).Where("user_id = ?", id)

    var total int64
    if err := q.Count(&total).Error; err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    var list []model.Commission
    if err := q.Order("id desc").Limit(limit).Offset((page-1)*limit).Find(&list).Error; err != nil {
        response.Error(c, http.StatusInternalServerError, err.Error())
        return
    }

    response.SuccessWithPagination(c, list, total, page, limit)
}
