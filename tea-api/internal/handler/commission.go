package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"tea-api/internal/model"
	"tea-api/internal/service/commission"
	"tea-api/pkg/database"
)

// CommissionHandler 佣金管理处理器
type CommissionHandler struct{}

// NewCommissionHandler 创建佣金处理器
func NewCommissionHandler() *CommissionHandler {
	return &CommissionHandler{}
}

// CalculateCommissionRequest 佣金计算请求
type CalculateCommissionRequest struct {
	OrderID          string  `json:"order_id" binding:"required"`
	PayerUserID      int64   `json:"payer_user_id" binding:"required"`
	ReferrerUserID   *int64  `json:"referrer_user_id"`
	Items            []OrderItemDTO `json:"items" binding:"required"`
	ShippingCents    int64   `json:"shipping_cents"`
	CouponCents      int64   `json:"coupon_cents"`
	DiscountCents    int64   `json:"discount_cents"`
}

// OrderItemDTO 订单项DTO
type OrderItemDTO struct {
	SKUID          string `json:"sku_id"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	Quantity       int    `json:"quantity"`
	DiscountCents  int64  `json:"discount_cents"`
}

// Calculate 计算佣金（预览，不写库）
// POST /api/v1/commissions/calculate
func (h *CommissionHandler) Calculate(c *gin.Context) {
	var req CalculateCommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	// 如果没有推荐人，返回空
	if req.ReferrerUserID == nil || *req.ReferrerUserID == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"message": "success",
			"data": gin.H{
				"commissions": []interface{}{},
				"total_amount_cents": 0,
			},
		})
		return
	}

	// 构建订单对象
	order := commission.Order{
		ID:             0, // 预览时不需要真实订单ID
		UserID:         req.PayerUserID,
		TotalAmount:    calculateTotalAmount(req.Items),
		ShippingAmount: req.ShippingCents,
		CouponAmount:   req.CouponCents,
		DiscountAmount: req.DiscountCents,
	}

	// 获取推荐人的佣金率（从数据库查询用户等级对应的佣金率）
	directRate, indirectRate := h.getCommissionRates(*req.ReferrerUserID)

	// 计算佣金
	records := commission.BuildCommissionRecords(order, *req.ReferrerUserID, directRate, indirectRate, 7)

	// 转换为响应格式
	var commissionList []gin.H
	totalAmount := int64(0)
	for _, record := range records {
		commissionList = append(commissionList, gin.H{
			"user_id":           record.UserID,
			"commission_type":   record.CommissionType,
			"rate":              record.Rate,
			"calculation_basis": record.CalculationBasis,
			"gross_amount":      record.GrossAmount,
			"net_amount":        record.NetAmount,
			"available_at":      record.AvailableAt.Format(time.RFC3339),
		})
		totalAmount += record.NetAmount
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"commissions":        commissionList,
			"total_amount_cents": totalAmount,
		},
	})
}

// CreateCommissions 持久化佣金记录（订单完成后调用）
// POST /api/v1/commissions
func (h *CommissionHandler) CreateCommissions(c *gin.Context) {
	var req CalculateCommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": err.Error()})
		return
	}

	// 检查该订单是否已经生成过佣金
	db := database.GetDB()
	var existingCount int64
	db.Model(&model.Commission{}).Where("order_id = ?", req.OrderID).Count(&existingCount)
	if existingCount > 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"message": "commissions already created",
			"data": gin.H{
				"created_count": existingCount,
			},
		})
		return
	}

	// 如果没有推荐人，不生成佣金
	if req.ReferrerUserID == nil || *req.ReferrerUserID == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"message": "no referrer, skip commission",
			"data": gin.H{
				"created_count": 0,
			},
		})
		return
	}

	// 构建订单对象
	order := commission.Order{
		ID:             0,
		UserID:         req.PayerUserID,
		TotalAmount:    calculateTotalAmount(req.Items),
		ShippingAmount: req.ShippingCents,
		CouponAmount:   req.CouponCents,
		DiscountAmount: req.DiscountCents,
	}

	// 获取推荐人的佣金率
	directRate, indirectRate := h.getCommissionRates(*req.ReferrerUserID)

	// 查找推荐人的上级（用于间接佣金）
	var ancestor *model.ReferralClosure
	db.Where("descendant_user_id = ? AND depth = 1", *req.ReferrerUserID).First(&ancestor)
	
	// 计算佣金
	records := commission.BuildCommissionRecords(order, *req.ReferrerUserID, directRate, indirectRate, 7)
	
	// 如果有上级，设置间接佣金的接收者
	if ancestor != nil {
		for i := range records {
			if records[i].CommissionType == "indirect" {
				records[i].UserID = int64(ancestor.AncestorUserID)
			}
		}
	} else {
		// 没有上级，移除间接佣金
		filtered := []commission.CommissionRecord{}
		for _, r := range records {
			if r.CommissionType != "indirect" {
				filtered = append(filtered, r)
			}
		}
		records = filtered
	}

	// 持久化佣金记录
	if err := commission.SaveCommissionRecords(records); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"created_count": len(records),
		},
	})
}

// ListUserCommissions 查询用户佣金列表
// GET /api/v1/users/:user_id/commissions
func (h *CommissionHandler) ListUserCommissions(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "invalid user_id"})
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	status := c.Query("status") // frozen/available/paid

	db := database.GetDB()
	query := db.Model(&model.Commission{}).Where("user_id = ?", userID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	var total int64
	query.Count(&total)

	// 查询列表
	var commissions []model.Commission
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&commissions)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"items": commissions,
			"total": total,
			"page":  page,
			"size":  pageSize,
		},
	})
}

// UnfreezeCommissions 解冻佣金（定时任务或管理员触发）
// POST /api/v1/admin/commissions/unfreeze
func (h *CommissionHandler) UnfreezeCommissions(c *gin.Context) {
	db := database.GetDB()
	now := time.Now()

	// 查找所有已到解冻时间且状态为frozen的佣金
	var commissions []model.Commission
	db.Where("status = ? AND available_at <= ?", "frozen", now).Find(&commissions)

	if len(commissions) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"message": "no commissions to unfreeze",
			"data": gin.H{
				"unfrozen_count": 0,
			},
		})
		return
	}

	// 批量更新状态
	unfrozenCount := 0
	tx := db.Begin()
	for _, comm := range commissions {
		if err := tx.Model(&comm).Update("status", "available").Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
			return
		}

		// 记录解冻流水
		cTx := model.CommissionTransaction{
			CommissionID: comm.ID,
			Type:         "unfreeze",
			Amount:       comm.NetAmount,
			BalanceAfter: comm.NetAmount,
		}
		if err := tx.Create(&cTx).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
			return
		}
		unfrozenCount++
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"unfrozen_count": unfrozenCount,
		},
	})
}

// getCommissionRates 获取用户的佣金率
func (h *CommissionHandler) getCommissionRates(userID int64) (directRate float64, indirectRate float64) {
	db := database.GetDB()
	
	// 查询用户信息获取partner_level_id
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		// 默认返回普通用户的佣金率
		return 0.10, 0.10
	}

	// 如果有合伙人等级，查询等级对应的佣金率
	if user.PartnerLevelID != nil && *user.PartnerLevelID > 0 {
		var level model.PartnerLevel
		if err := db.First(&level, *user.PartnerLevelID).Error; err == nil {
			directRate, _ = level.DirectCommissionRate.Float64()
			indirectRate, _ = level.TeamCommissionRate.Float64()
			return directRate, indirectRate
		}
	}

	// 如果有会员等级，查询会员套餐对应的佣金率
	if user.MembershipPackageID != nil && *user.MembershipPackageID > 0 {
		var pkg model.MembershipPackage
		if err := db.First(&pkg, *user.MembershipPackageID).Error; err == nil {
			directRate, _ = pkg.DirectCommissionRate.Float64()
			indirectRate, _ = pkg.TeamCommissionRate.Float64()
			return directRate, indirectRate
		}
	}

	// 默认佣金率
	return 0.10, 0.10
}

// calculateTotalAmount 计算订单项总金额
func calculateTotalAmount(items []OrderItemDTO) int64 {
	total := int64(0)
	for _, item := range items {
		total += item.UnitPriceCents * int64(item.Quantity)
	}
	return total
}

// GetCommissionSummary 获取用户佣金汇总
// GET /api/v1/users/:user_id/commissions/summary
func (h *CommissionHandler) GetCommissionSummary(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4000, "message": "invalid user_id"})
		return
	}

	db := database.GetDB()

	// 统计各状态佣金总额
	var frozenSum, availableSum, paidSum decimal.Decimal
	
	db.Model(&model.Commission{}).
		Where("user_id = ? AND status = ?", userID, "frozen").
		Select("COALESCE(SUM(net_amount), 0)").
		Scan(&frozenSum)
	
	db.Model(&model.Commission{}).
		Where("user_id = ? AND status = ?", userID, "available").
		Select("COALESCE(SUM(net_amount), 0)").
		Scan(&availableSum)
	
	db.Model(&model.Commission{}).
		Where("user_id = ? AND status = ?", userID, "paid").
		Select("COALESCE(SUM(net_amount), 0)").
		Scan(&paidSum)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
		"data": gin.H{
			"frozen_amount":    frozenSum,
			"available_amount": availableSum,
			"paid_amount":      paidSum,
			"total_amount":     frozenSum.Add(availableSum).Add(paidSum),
		},
	})
}
