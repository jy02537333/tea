package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"

    "tea-api/internal/model"
    "tea-api/pkg/database"
    "tea-api/pkg/utils"
)

// SC2: 用户侧推荐/被推荐接口（最小联通）

type ReferralHandler struct{}

func NewReferralHandler() *ReferralHandler { return &ReferralHandler{} }

type referralRecordReq struct {
    ReferrerID      uint   `json:"referrer_id" binding:"required"`
    ReferredUserID  uint   `json:"referred_user_id" binding:"required"`
    Source          string `json:"source"`
}

// POST /api/v1/referral/record
func (h *ReferralHandler) Record(c *gin.Context) {
    var req referralRecordReq
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.Error(c, utils.CodeInvalidParam, "参数错误")
        return
    }

    db := database.GetDB()
    var existing model.ReferralClosure
    if err := db.Where("ancestor_user_id = ? AND descendant_user_id = ?", req.ReferrerID, req.ReferredUserID).First(&existing).Error; err == nil {
        utils.Success(c, gin.H{"created": false, "note": "exists"})
        return
    }

    rc := model.ReferralClosure{AncestorUserID: req.ReferrerID, DescendantUserID: req.ReferredUserID, Depth: 1}
    if err := db.Create(&rc).Error; err != nil {
        utils.Error(c, utils.CodeError, err.Error())
        return
    }
    utils.Success(c, gin.H{"created": true})
}

// GET /api/v1/users/:id/referral-stats
func (h *ReferralHandler) Stats(c *gin.Context) {
    uid := parseUintParamRef(c, "id")
    if uid == 0 { utils.Error(c, utils.CodeInvalidParam, "invalid user id"); return }

    db := database.GetDB()
    var direct int64
    var team int64
    db.Model(&model.ReferralClosure{}).Where("ancestor_user_id = ? AND depth = 1", uid).Count(&direct)
    db.Model(&model.ReferralClosure{}).Where("ancestor_user_id = ? AND depth > 1", uid).Count(&team)
    utils.Success(c, gin.H{"user_id": uid, "direct_count": direct, "team_count": team})
}

// GET /api/v1/users/:id/referred-users?page=&size=
func (h *ReferralHandler) ListReferredUsers(c *gin.Context) {
    uid := parseUintParamRef(c, "id")
    if uid == 0 { utils.Error(c, utils.CodeInvalidParam, "invalid user id"); return }

    page := parseIntQueryRef(c, "page", 1)
    size := parseIntQueryRef(c, "size", 20)
    if size > 100 { size = 100 }
    offset := (page - 1) * size

    db := database.GetDB()
    var total int64
    db.Model(&model.ReferralClosure{}).Where("ancestor_user_id = ?", uid).Count(&total)

    var rows []model.ReferralClosure
    if err := db.Where("ancestor_user_id = ?", uid).Order("depth asc, descendant_user_id asc").Limit(size).Offset(offset).Find(&rows).Error; err != nil {
        utils.Error(c, utils.CodeError, err.Error())
        return
    }
    // 简化输出结构
    out := make([]gin.H, 0, len(rows))
    for _, r := range rows {
        out = append(out, gin.H{"user_id": r.DescendantUserID, "depth": r.Depth})
    }
    c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"total": total, "list": out}})
}

func parseUintParamRef(c *gin.Context, key string) uint {
    v := c.Param(key)
    if v == "" { return 0 }
    x, err := strconv.ParseUint(v, 10, 64)
    if err != nil { return 0 }
    return uint(x)
}

func parseIntQueryRef(c *gin.Context, key string, dft int) int {
    v := c.Query(key)
    if v == "" { return dft }
    x, err := strconv.Atoi(v)
    if err != nil { return dft }
    return x
}
