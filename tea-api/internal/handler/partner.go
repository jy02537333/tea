package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/pkg/database"
)

// PartnerHandler 合伙人/会员处理器
type PartnerHandler struct{}

// NewPartnerHandler 创建合伙人处理器
func NewPartnerHandler() *PartnerHandler {
	return &PartnerHandler{}
}

// ListPackages 获取会员/合伙人礼包列表
// GET /api/v1/partner/packages
func (h *PartnerHandler) ListPackages(c *gin.Context) {
	packageType := c.Query("type") // membership/partner_package
	
	db := database.GetDB()
	query := db.Model(&model.MembershipPackage{})
	
	if packageType != "" {
		query = query.Where("type = ?", packageType)
	}

	var packages []model.MembershipPackage
	query.Order("price ASC").Find(&packages)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"items": packages,
		},
	})
}

// PurchasePackageRequest 购买礼包请求
type PurchasePackageRequest struct {
	UserID    uint   `json:"user_id" binding:"required"`
	PackageID uint   `json:"package_id" binding:"required"`
	PayMethod string `json:"pay_method"` // wechat/alipay/balance
}

// PurchasePackage 购买会员/合伙人礼包
// POST /api/v1/membership/purchase
func (h *PartnerHandler) PurchasePackage(c *gin.Context) {
	var req PurchasePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	db := database.GetDB()

	// 查询礼包信息
	var pkg model.MembershipPackage
	if err := db.First(&pkg, req.PackageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 4004, "message": "package not found"})
		return
	}

	// 查询用户信息
	var user model.User
	if err := db.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 4004, "message": "user not found"})
		return
	}

	// 创建会员礼包订单
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建订单（简化版，实际应该使用完整的订单流程）
	order := model.Order{
		UserID:      req.UserID,
		OrderNo:     generateOrderNo(),
		OrderType:   1, // 1:商城
		Status:      1, // 1:待付款
		TotalAmount: pkg.Price,
		PayAmount:   pkg.Price,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	// 这里应该调用支付接口，简化处理，直接返回订单信息
	// 实际场景中，支付成功后的回调会触发升级和权益发放

	c.JSON(http.StatusCreated, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"order_id":   order.ID,
			"order_no":   order.OrderNo,
			"amount":     pkg.Price,
			"pay_method": req.PayMethod,
		},
	})
}

// UpgradePartner 升级为合伙人（支付成功后的回调处理）
// POST /api/v1/partner/upgrade
func (h *PartnerHandler) UpgradePartner(c *gin.Context) {
	type UpgradeRequest struct {
		UserID    uint `json:"user_id" binding:"required"`
		PackageID uint `json:"package_id" binding:"required"`
		OrderID   uint `json:"order_id" binding:"required"`
	}

	var req UpgradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	db := database.GetDB()

	// 查询礼包信息
	var pkg model.MembershipPackage
	if err := db.First(&pkg, req.PackageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 4004, "message": "package not found"})
		return
	}

	// 查询用户信息
	var user model.User
	if err := db.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 4004, "message": "user not found"})
		return
	}

	// 检查是否已经升级过
	if user.MembershipPackageID != nil && *user.MembershipPackageID >= req.PackageID {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"message": "already upgraded",
		})
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新用户等级
	updates := map[string]interface{}{
		"membership_package_id": req.PackageID,
	}

	// 如果是合伙人礼包，查找对应的 partner_level
	// 优先通过package关联字段，退而求其次使用名称匹配
	if pkg.Type == "partner_package" {
		var partnerLevel model.PartnerLevel
		// 尝试使用名称匹配查找等级
		err := tx.Where("name = ? OR name LIKE ?", pkg.Name, pkg.Name+"%").First(&partnerLevel).Error
		if err == nil {
			levelID := partnerLevel.ID
			updates["partner_level_id"] = &levelID
		}
		// 如果找不到匹配的等级，不设置partner_level_id（保持为NULL）
	}

	if err := tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	// 发放茶币奖励（如果有）
	if pkg.TeaCoinAward.GreaterThan(decimal.Zero) {
		var wallet model.Wallet
		if err := tx.Where("user_id = ?", req.UserID).First(&wallet).Error; err != nil {
			// 如果钱包不存在，创建一个
			wallet = model.Wallet{
				UserID:   req.UserID,
				Balance:  decimal.Zero,
				TeaCoins: pkg.TeaCoinAward,
			}
			tx.Create(&wallet)
		} else {
			// 更新茶币余额
			tx.Model(&wallet).Update("tea_coins", wallet.TeaCoins.Add(pkg.TeaCoinAward))
		}

		// 记录流水
		walletTx := model.WalletTransaction{
			UserID:       req.UserID,
			Type:         "tea_coin_award",
			Amount:       pkg.TeaCoinAward,
			BalanceAfter: wallet.Balance,
			OrderID:      &req.OrderID,
			Description:  "购买" + pkg.Name + "赠送茶币",
		}
		tx.Create(&walletTx)
	}

	// 如果用户有推荐人，计算升级奖励
	var referral model.Referral
	if err := tx.Where("referred_user_id = ?", req.UserID).First(&referral).Error; err == nil {
		// 查询推荐人的旧等级
		var referrer model.User
		tx.First(&referrer, referral.ReferrerUserID)

		// 计算升级奖励（基于差价）
		if pkg.UpgradeRewardRate.GreaterThan(decimal.Zero) {
			// 简化处理：按当前礼包价格计算
			rewardAmount := pkg.Price.Mul(pkg.UpgradeRewardRate)

			// 创建升级奖励佣金记录
			commission := model.Commission{
				UserID:           referral.ReferrerUserID,
				OrderID:          &req.OrderID,
				PackageID:        &req.PackageID,
				CommissionType:   "upgrade",
				SourceUserID:     &req.UserID,
				Rate:             pkg.UpgradeRewardRate,
				CalculationBasis: pkg.Price,
				GrossAmount:      rewardAmount,
				Fee:              decimal.Zero,
				NetAmount:        rewardAmount,
				Status:           "frozen",
			}

			availableAt := time.Now().Add(7 * 24 * time.Hour)
			commission.AvailableAt = &availableAt

			if err := tx.Create(&commission).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "upgrade success",
		"data": gin.H{
			"package_id":     req.PackageID,
			"tea_coin_award": pkg.TeaCoinAward,
		},
	})
}

// ListPartnerLevels 获取合伙人等级列表（管理端）
// GET /api/v1/admin/partner-levels
func (h *PartnerHandler) ListPartnerLevels(c *gin.Context) {
	db := database.GetDB()

	var levels []model.PartnerLevel
	db.Order("id ASC").Find(&levels)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"items": levels,
		},
	})
}

// CreatePartnerLevel 创建合伙人等级（管理端）
// POST /api/v1/admin/partner-levels
func (h *PartnerHandler) CreatePartnerLevel(c *gin.Context) {
	var level model.PartnerLevel
	if err := c.ShouldBindJSON(&level); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	db := database.GetDB()
	if err := db.Create(&level).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 0,
		"message": "success",
		"data": level,
	})
}

// UpdatePartnerLevel 更新合伙人等级（管理端）
// PUT /api/v1/admin/partner-levels/:id
func (h *PartnerHandler) UpdatePartnerLevel(c *gin.Context) {
	id := c.Param("id")
	
	var level model.PartnerLevel
	if err := c.ShouldBindJSON(&level); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	db := database.GetDB()
	if err := db.Model(&model.PartnerLevel{}).Where("id = ?", id).Updates(&level).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
	})
}

// generateOrderNo 生成订单号（简化版）
func generateOrderNo() string {
	return "MO" + time.Now().Format("20060102150405") + strconv.FormatInt(time.Now().UnixNano()%10000, 10)
}
