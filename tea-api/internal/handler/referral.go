package handler

import (
	"net/http"

	"tea-api/internal/model"
	"tea-api/pkg/database"

	"github.com/gin-gonic/gin"
)

type bindReferralReq struct {
    ReferrerID uint `json:"referrer_id" binding:"required"`
}

// BindReferral 建立推荐关系（最小版：插入 referrals_closure depth=1）
func BindReferral(c *gin.Context) {
    uidVal, _ := c.Get("user_id")
    currUserID := uint(uidVal.(uint))

    var req bindReferralReq
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 4001, "message": "参数错误", "data": nil})
        return
    }
    if req.ReferrerID == 0 || req.ReferrerID == currUserID {
        c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"bound": false}})
        return
    }

    db := database.GetDB()
    // 简化：若已存在则跳过；否则插入一条闭包关系（referrer -> current user, depth=1）
    var existing model.ReferralClosure
    if err := db.Where("ancestor_user_id = ? AND descendant_user_id = ? AND depth = 1", req.ReferrerID, currUserID).First(&existing).Error; err == nil && existing.AncestorUserID != 0 {
        c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"bound": true}})
        return
    }

    rc := model.ReferralClosure{AncestorUserID: req.ReferrerID, DescendantUserID: currUserID, Depth: 1}
    if err := db.Create(&rc).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 5001, "message": "绑定失败", "data": nil})
        return
    }
    // 可选：确保当前用户的自闭包（depth=0）存在（忽略错误）
    _ = db.Where("ancestor_user_id = ? AND descendant_user_id = ? AND depth = 0", currUserID, currUserID).First(&existing).Error
    if existing.AncestorUserID == 0 {
        _ = db.Create(&model.ReferralClosure{AncestorUserID: currUserID, DescendantUserID: currUserID, Depth: 0}).Error
    }

    c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"bound": true}})
}
