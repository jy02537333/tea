package handler

import (
	"errors"
	"net/http"

	"tea-api/internal/model"
	"tea-api/pkg/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
    // 规则：用户多次点击不同分享人的链接，以最后一次点击为准覆盖直推关系（descendant depth=1）。
    // 注意：当前仅维护 depth=0/1 的最小闭包，不做全量祖先链重算。
    if err := db.Transaction(func(tx *gorm.DB) error {
        // 校验推荐人存在
        var refUser model.User
        if err := tx.Select("id").First(&refUser, req.ReferrerID).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return errors.New("referrer_not_found")
            }
            return err
        }

        // 先看是否已绑定为同一推荐人
        var existing model.ReferralClosure
        err := tx.Where("descendant_user_id = ? AND depth = 1", currUserID).First(&existing).Error
        if err == nil {
            if existing.AncestorUserID == req.ReferrerID {
                return nil
            }
            // 覆盖：删掉当前用户所有 depth=1 直推，再插入新的
            if err := tx.Where("descendant_user_id = ? AND depth = 1", currUserID).Delete(&model.ReferralClosure{}).Error; err != nil {
                return err
            }
        } else if !errors.Is(err, gorm.ErrRecordNotFound) {
            return err
        }

        // 确保当前用户自闭包 depth=0 存在（忽略并发/重复）
        var self model.ReferralClosure
        _ = tx.Where("ancestor_user_id = ? AND descendant_user_id = ? AND depth = 0", currUserID, currUserID).First(&self).Error
        if self.AncestorUserID == 0 {
            _ = tx.Create(&model.ReferralClosure{AncestorUserID: currUserID, DescendantUserID: currUserID, Depth: 0}).Error
        }

        rc := model.ReferralClosure{AncestorUserID: req.ReferrerID, DescendantUserID: currUserID, Depth: 1}
        if err := tx.Create(&rc).Error; err != nil {
            return err
        }
        return nil
    }); err != nil {
        if err.Error() == "referrer_not_found" {
            c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"bound": false}})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"code": 5001, "message": "绑定失败", "data": nil})
        return
    }

    c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"bound": true}})
}
