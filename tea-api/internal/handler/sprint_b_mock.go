package handler

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"

    "tea-api/pkg/utils"
)

// UsersMeSummaryMock 返回个人中心聚合（占位Mock）
// GET /api/v1/users/me/summary
func UsersMeSummaryMock(c *gin.Context) {
    uid := uint(0)
    if v, ok := c.Get("user_id"); ok {
        if vv, ok2 := v.(uint); ok2 {
            uid = vv
        }
    }
    utils.Success(c, gin.H{
        "user": gin.H{
            "id":           uid,
            "name":         "Mock 用户",
            "avatar":       "",
            "member_level": "basic",
        },
        "wallet": gin.H{
            "balance_cents": 12345,
            "tea_coin":      50,
        },
        "points": gin.H{
            "balance": 1200,
        },
        "coupons": gin.H{
            "available": 2,
            "expiring":  1,
        },
        "server_time": time.Now().Format(time.RFC3339),
        "mock":        true,
    })
}

// WalletSummaryMock 返回钱包概览（占位Mock）
// GET /api/v1/wallet
func WalletSummaryMock(c *gin.Context) {
    utils.Success(c, gin.H{
        "balance_cents": 12345,
        "tea_coin":      50,
        "transactions": []gin.H{},
        "mock":          true,
    })
}

// PointsSummaryMock 返回积分概览（占位Mock）
// GET /api/v1/points
func PointsSummaryMock(c *gin.Context) {
    utils.Success(c, gin.H{
        "balance":      1200,
        "recent_flows": []gin.H{},
        "mock":         true,
    })
}

// 兜底：在 Mock 处理器中也提供一个简易健康检查（不注册路由）
func mockOK(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }
