package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

// ReferralHandler 推荐关系处理器
type ReferralHandler struct{}

// NewReferralHandler 创建推荐关系处理器
func NewReferralHandler() *ReferralHandler {
	return &ReferralHandler{}
}

// RecordReferralRequest 记录推荐关系请求
type RecordReferralRequest struct {
	ReferrerUserID  uint   `json:"referrer_user_id" binding:"required"`
	ReferredUserID  uint   `json:"referred_user_id" binding:"required"`
	Source          string `json:"source"` // 来源：share_link/qrcode/invite_code
}

// RecordReferral 记录推荐关系
// POST /api/v1/referral/record
func (h *ReferralHandler) RecordReferral(c *gin.Context) {
	var req RecordReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	db := database.GetDB()

	// 检查被推荐人是否已有推荐关系
	var existingReferral model.Referral
	if err := db.Where("referred_user_id = ?", req.ReferredUserID).First(&existingReferral).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"message": "referral relationship already exists",
			"data": gin.H{
				"referrer_user_id": existingReferral.ReferrerUserID,
			},
		})
		return
	}

	// 防止自己推荐自己
	if req.ReferrerUserID == req.ReferredUserID {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "cannot refer yourself"})
		return
	}

	// 创建推荐关系
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	referral := model.Referral{
		ReferrerUserID:  req.ReferrerUserID,
		ReferredUserID:  req.ReferredUserID,
		Source:          req.Source,
	}

	if err := tx.Create(&referral).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	// 构建闭包表记录（自己到自己 depth=0）
	selfClosure := model.ReferralClosure{
		AncestorUserID:   req.ReferredUserID,
		DescendantUserID: req.ReferredUserID,
		Depth:            0,
	}
	if err := tx.Create(&selfClosure).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	// 查找推荐人的所有祖先，为被推荐人建立与所有祖先的闭包关系
	var ancestorClosures []model.ReferralClosure
	tx.Where("descendant_user_id = ?", req.ReferrerUserID).Find(&ancestorClosures)

	for _, ac := range ancestorClosures {
		newClosure := model.ReferralClosure{
			AncestorUserID:   ac.AncestorUserID,
			DescendantUserID: req.ReferredUserID,
			Depth:            ac.Depth + 1,
		}
		if err := tx.Create(&newClosure).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"referral_id": referral.ID,
		},
	})
}

// GetReferralInfo 获取用户的推荐信息
// GET /api/v1/users/:user_id/referral-info
func (h *ReferralHandler) GetReferralInfo(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "invalid user_id"})
		return
	}

	db := database.GetDB()

	// 查询推荐人
	var referral model.Referral
	referrerID := uint(0)
	if err := db.Where("referred_user_id = ?", userID).First(&referral).Error; err == nil {
		referrerID = referral.ReferrerUserID
	}

	// 统计直推人数
	var directCount int64
	db.Model(&model.Referral{}).Where("referrer_user_id = ?", userID).Count(&directCount)

	// 统计团队总人数（通过闭包表）
	var teamCount int64
	db.Model(&model.ReferralClosure{}).
		Where("ancestor_user_id = ? AND depth > 0", userID).
		Count(&teamCount)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"user_id":       userID,
			"referrer_id":   referrerID,
			"direct_count":  directCount,
			"team_count":    teamCount,
		},
	})
}

// ListReferredUsers 获取用户的直推列表
// GET /api/v1/users/:user_id/referred-users
func (h *ReferralHandler) ListReferredUsers(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "invalid user_id"})
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	db := database.GetDB()

	// 统计总数
	var total int64
	db.Model(&model.Referral{}).Where("referrer_user_id = ?", userID).Count(&total)

	// 查询列表
	var referrals []model.Referral
	offset := (page - 1) * pageSize
	db.Where("referrer_user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&referrals)

	// 获取被推荐人的详细信息
	type ReferredUserInfo struct {
		UserID       uint   `json:"user_id"`
		Phone        string `json:"phone"`
		Nickname     string `json:"nickname"`
		ReferredAt   string `json:"referred_at"`
		Source       string `json:"source"`
	}

	var userInfos []ReferredUserInfo
	for _, ref := range referrals {
		var user model.User
		var profile model.UserProfile
		
		if err := db.First(&user, ref.ReferredUserID).Error; err != nil {
			// 用户不存在，跳过
			continue
		}
		
		// profile 可能不存在，使用默认值
		db.Where("user_id = ?", ref.ReferredUserID).First(&profile)

		userInfos = append(userInfos, ReferredUserInfo{
			UserID:     ref.ReferredUserID,
			Phone:      user.Phone,
			Nickname:   profile.Nickname,
			ReferredAt: ref.CreatedAt.Format("2006-01-02 15:04:05"),
			Source:     ref.Source,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"items": userInfos,
			"total": total,
			"page":  page,
			"size":  pageSize,
		},
	})
}

// GetReferralStats 获取推荐统计数据（用于分享页面）
// GET /api/v1/users/:user_id/referral-stats
func (h *ReferralHandler) GetReferralStats(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "invalid user_id"})
		return
	}

	db := database.GetDB()

	// 统计直推人数
	var directCount int64
	db.Model(&model.Referral{}).Where("referrer_user_id = ?", userID).Count(&directCount)

	// 统计团队总人数
	var teamCount int64
	db.Model(&model.ReferralClosure{}).
		Where("ancestor_user_id = ? AND depth > 0", userID).
		Count(&teamCount)

	// 统计累计佣金
	var totalCommission, availableCommission interface{}
	db.Model(&model.Commission{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(net_amount), 0)").
		Scan(&totalCommission)

	db.Model(&model.Commission{}).
		Where("user_id = ? AND status = ?", userID, "available").
		Select("COALESCE(SUM(net_amount), 0)").
		Scan(&availableCommission)

	// 统计本月新增推荐人数
	var monthlyCount int64
	db.Model(&model.Referral{}).
		Where("referrer_user_id = ? AND MONTH(created_at) = MONTH(NOW()) AND YEAR(created_at) = YEAR(NOW())", userID).
		Count(&monthlyCount)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"direct_count":         directCount,
			"team_count":           teamCount,
			"total_commission":     totalCommission,
			"available_commission": availableCommission,
			"monthly_new_count":    monthlyCount,
		},
	})
}
