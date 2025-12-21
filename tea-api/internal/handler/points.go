package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"tea-api/pkg/database"
	"tea-api/pkg/response"
)

// GetMyPoints GET /api/v1/points
// 返回当前用户积分总额（可用）
func GetMyPoints(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未获取到用户身份")
		return
	}
	// 优先聚合 points_transactions
	type row struct{ Total int64 }
	var r row
	err := database.GetDB().Table("points_transactions").Select("COALESCE(SUM(`change`),0) AS total").Where("user_id = ?", uid).Take(&r).Error
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "积分查询失败")
		return
	}
	response.Success(c, gin.H{"points": r.Total})
}

// ListMyPointsTransactions GET /api/v1/points/transactions
// 返回积分流水记录（分页）
func ListMyPointsTransactions(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未获取到用户身份")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	type raw struct {
		ID        uint
		Change    int64
		Reason    string
		CreatedAt time.Time
	}
	var total int64
	q := database.GetDB().Table("points_transactions").Where("user_id = ?", uid)
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "统计失败")
		return
	}
	var raws []raw
	if err := q.Order("id desc").Limit(limit).Offset((page - 1) * limit).Find(&raws).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	items := make([]gin.H, 0, len(raws))
	for _, r := range raws {
		items = append(items, gin.H{
			"id":         r.ID,
			"change":     r.Change,
			"reason":     r.Reason,
			"created_at": r.CreatedAt.Format(time.RFC3339),
		})
	}
	response.SuccessWithPagination(c, items, total, page, limit)
}
