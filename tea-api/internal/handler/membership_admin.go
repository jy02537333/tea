package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/pkg/database"
	"tea-api/pkg/response"
)

// MembershipAdminHandler 系统层会员/合伙人配置管理
type MembershipAdminHandler struct{}

func NewMembershipAdminHandler() *MembershipAdminHandler { return &MembershipAdminHandler{} }

// ==== 会员套餐（MembershipPackage）====

// ListPackages GET /api/v1/admin/membership-packages
func (h *MembershipAdminHandler) ListPackages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}

	typeFilter := strings.TrimSpace(c.Query("type"))

	db := database.GetDB()
	q := db.Model(&model.MembershipPackage{})
	if typeFilter != "" {
		q = q.Where("type = ?", typeFilter)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	var list []model.MembershipPackage
	if err := q.Order("id desc").Limit(limit).Offset((page - 1) * limit).Find(&list).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithPagination(c, list, total, page, limit)
}

type membershipPackageCreateReq struct {
	Name                 string   `json:"name" binding:"required"`
	Price                float64  `json:"price" binding:"required"`
	TeaCoinAward         *float64 `json:"tea_coin_award"`
	DiscountRate         *float64 `json:"discount_rate"`
	PurchaseDiscountRate *float64 `json:"purchase_discount_rate"`
	DirectCommissionRate *float64 `json:"direct_commission_rate"`
	TeamCommissionRate   *float64 `json:"team_commission_rate"`
	UpgradeRewardRate    *float64 `json:"upgrade_reward_rate"`
	Type                 string   `json:"type"`
}

// CreatePackage POST /api/v1/admin/membership-packages
func (h *MembershipAdminHandler) CreatePackage(c *gin.Context) {
	var req membershipPackageCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		response.BadRequest(c, "名称不能为空")
		return
	}

	// 默认类型为 membership
	typeVal := strings.TrimSpace(req.Type)
	if typeVal == "" {
		typeVal = "membership"
	}

	pkg := &model.MembershipPackage{
		Name:  strings.TrimSpace(req.Name),
		Price: decimal.NewFromFloat(req.Price),
		Type:  typeVal,
	}
	if req.TeaCoinAward != nil {
		pkg.TeaCoinAward = decimal.NewFromFloat(*req.TeaCoinAward)
	}
	if req.DiscountRate != nil {
		pkg.DiscountRate = decimal.NewFromFloat(*req.DiscountRate)
	}
	if req.PurchaseDiscountRate != nil {
		pkg.PurchaseDiscountRate = decimal.NewFromFloat(*req.PurchaseDiscountRate)
	}
	if req.DirectCommissionRate != nil {
		pkg.DirectCommissionRate = decimal.NewFromFloat(*req.DirectCommissionRate)
	}
	if req.TeamCommissionRate != nil {
		pkg.TeamCommissionRate = decimal.NewFromFloat(*req.TeamCommissionRate)
	}
	if req.UpgradeRewardRate != nil {
		pkg.UpgradeRewardRate = decimal.NewFromFloat(*req.UpgradeRewardRate)
	}

	if err := database.GetDB().Create(pkg).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, pkg)
}

type membershipPackageUpdateReq struct {
	Name                 *string  `json:"name"`
	Price                *float64 `json:"price"`
	TeaCoinAward         *float64 `json:"tea_coin_award"`
	DiscountRate         *float64 `json:"discount_rate"`
	PurchaseDiscountRate *float64 `json:"purchase_discount_rate"`
	DirectCommissionRate *float64 `json:"direct_commission_rate"`
	TeamCommissionRate   *float64 `json:"team_commission_rate"`
	UpgradeRewardRate    *float64 `json:"upgrade_reward_rate"`
	Type                 *string  `json:"type"`
}

// UpdatePackage PUT /api/v1/admin/membership-packages/:id
func (h *MembershipAdminHandler) UpdatePackage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "非法ID")
		return
	}

	var req membershipPackageUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	db := database.GetDB()
	var pkg model.MembershipPackage
	if err := db.First(&pkg, id).Error; err != nil {
		response.NotFound(c, "记录不存在")
		return
	}

	updates := map[string]any{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			response.BadRequest(c, "名称不能为空")
			return
		}
		updates["name"] = name
	}
	if req.Price != nil {
		updates["price"] = decimal.NewFromFloat(*req.Price)
	}
	if req.TeaCoinAward != nil {
		updates["tea_coin_award"] = decimal.NewFromFloat(*req.TeaCoinAward)
	}
	if req.DiscountRate != nil {
		updates["discount_rate"] = decimal.NewFromFloat(*req.DiscountRate)
	}
	if req.PurchaseDiscountRate != nil {
		updates["purchase_discount_rate"] = decimal.NewFromFloat(*req.PurchaseDiscountRate)
	}
	if req.DirectCommissionRate != nil {
		updates["direct_commission_rate"] = decimal.NewFromFloat(*req.DirectCommissionRate)
	}
	if req.TeamCommissionRate != nil {
		updates["team_commission_rate"] = decimal.NewFromFloat(*req.TeamCommissionRate)
	}
	if req.UpgradeRewardRate != nil {
		updates["upgrade_reward_rate"] = decimal.NewFromFloat(*req.UpgradeRewardRate)
	}
	if req.Type != nil {
		updates["type"] = strings.TrimSpace(*req.Type)
	}

	if len(updates) == 0 {
		response.Success(c, pkg)
		return
	}

	if err := db.Model(&pkg).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, pkg)
}

// DeletePackage DELETE /api/v1/admin/membership-packages/:id
func (h *MembershipAdminHandler) DeletePackage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "非法ID")
		return
	}

	if err := database.GetDB().Delete(&model.MembershipPackage{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"ok": true})
}

// ==== 合伙人等级（PartnerLevel）====

// ListPartnerLevels GET /api/v1/admin/partner-levels
func (h *MembershipAdminHandler) ListPartnerLevels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	db := database.GetDB()
	q := db.Model(&model.PartnerLevel{})

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	var list []model.PartnerLevel
	if err := q.Order("id asc").Limit(limit).Offset((page - 1) * limit).Find(&list).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithPagination(c, list, total, page, limit)
}

type partnerLevelCreateReq struct {
	Name                 string   `json:"name" binding:"required"`
	PurchaseDiscountRate *float64 `json:"purchase_discount_rate"`
	DirectCommissionRate *float64 `json:"direct_commission_rate"`
	TeamCommissionRate   *float64 `json:"team_commission_rate"`
	UpgradeRewardRate    *float64 `json:"upgrade_reward_rate"`
	Note                 string   `json:"note"`
}

// CreatePartnerLevel POST /api/v1/admin/partner-levels
func (h *MembershipAdminHandler) CreatePartnerLevel(c *gin.Context) {
	var req partnerLevelCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		response.BadRequest(c, "名称不能为空")
		return
	}

	level := &model.PartnerLevel{
		Name: strings.TrimSpace(req.Name),
		Note: strings.TrimSpace(req.Note),
	}
	if req.PurchaseDiscountRate != nil {
		level.PurchaseDiscountRate = decimal.NewFromFloat(*req.PurchaseDiscountRate)
	}
	if req.DirectCommissionRate != nil {
		level.DirectCommissionRate = decimal.NewFromFloat(*req.DirectCommissionRate)
	}
	if req.TeamCommissionRate != nil {
		level.TeamCommissionRate = decimal.NewFromFloat(*req.TeamCommissionRate)
	}
	if req.UpgradeRewardRate != nil {
		level.UpgradeRewardRate = decimal.NewFromFloat(*req.UpgradeRewardRate)
	}

	if err := database.GetDB().Create(level).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, level)
}

type partnerLevelUpdateReq struct {
	Name                 *string  `json:"name"`
	PurchaseDiscountRate *float64 `json:"purchase_discount_rate"`
	DirectCommissionRate *float64 `json:"direct_commission_rate"`
	TeamCommissionRate   *float64 `json:"team_commission_rate"`
	UpgradeRewardRate    *float64 `json:"upgrade_reward_rate"`
	Note                 *string  `json:"note"`
}

// UpdatePartnerLevel PUT /api/v1/admin/partner-levels/:id
func (h *MembershipAdminHandler) UpdatePartnerLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "非法ID")
		return
	}

	var req partnerLevelUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	db := database.GetDB()
	var level model.PartnerLevel
	if err := db.First(&level, id).Error; err != nil {
		response.NotFound(c, "记录不存在")
		return
	}

	updates := map[string]any{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			response.BadRequest(c, "名称不能为空")
			return
		}
		updates["name"] = name
	}
	if req.PurchaseDiscountRate != nil {
		updates["purchase_discount_rate"] = decimal.NewFromFloat(*req.PurchaseDiscountRate)
	}
	if req.DirectCommissionRate != nil {
		updates["direct_commission_rate"] = decimal.NewFromFloat(*req.DirectCommissionRate)
	}
	if req.TeamCommissionRate != nil {
		updates["team_commission_rate"] = decimal.NewFromFloat(*req.TeamCommissionRate)
	}
	if req.UpgradeRewardRate != nil {
		updates["upgrade_reward_rate"] = decimal.NewFromFloat(*req.UpgradeRewardRate)
	}
	if req.Note != nil {
		updates["note"] = strings.TrimSpace(*req.Note)
	}

	if len(updates) == 0 {
		response.Success(c, level)
		return
	}

	if err := db.Model(&level).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, level)
}

// DeletePartnerLevel DELETE /api/v1/admin/partner-levels/:id
func (h *MembershipAdminHandler) DeletePartnerLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "非法ID")
		return
	}

	if err := database.GetDB().Delete(&model.PartnerLevel{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"ok": true})
}
