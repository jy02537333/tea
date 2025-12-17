package handler

import (
	"net/http"
	"strconv"

	mw "tea-api/internal/middleware"
	svc "tea-api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterUserSummaryRoutes registers routes for user profile summary.
func RegisterUserSummaryRoutes(r *gin.RouterGroup) {
	r.GET("/users/me/summary", mw.AuthJWT(), GetUserSummary)
}

// GetUserSummary returns aggregated profile data for current user.
// TODO: wire with service layer and auth middleware to fetch real data.
func GetUserSummary(c *gin.Context) {
	// Extract user id from auth middleware context
	var userID int64 = 0
	if v, exists := c.Get("user_id"); exists {
		switch t := v.(type) {
		case int64:
			userID = t
		case int:
			userID = int64(t)
		case uint:
			userID = int64(t)
		case string:
			if id, err := strconv.ParseInt(t, 10, 64); err == nil {
				userID = id
			}
		}
	}
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 4012, "message": "未获取到用户身份", "data": nil})
		return
	}
	s := svc.NewUserSummaryServiceFromGlobal()
	summary, err := s.GetSummary(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 5001, "message": "获取用户摘要失败", "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": summary})
}
