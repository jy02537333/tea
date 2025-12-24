package handler

import (
    "time"

    "github.com/gin-gonic/gin"

    "tea-api/pkg/utils"
)

// PrintAdminHandler 管理端打印任务占位实现
type PrintAdminHandler struct{}

func NewPrintAdminHandler() *PrintAdminHandler { return &PrintAdminHandler{} }

// CreateTask 创建打印任务（占位实现：返回生成的任务ID与排队状态，不做实际持久化）
// 路由：POST /api/v1/admin/print/tasks
func (h *PrintAdminHandler) CreateTask(c *gin.Context) {
    var req struct {
        OrderID   *uint  `json:"order_id"`
        Target    string `json:"target"`    // 打印目标（如：kitchen/receipt）
        Payload   string `json:"payload"`   // 打印内容（JSON/文本）
        Priority  int    `json:"priority"`  // 可选优先级
    }
    _ = c.ShouldBindJSON(&req)

    id := "pt_" + time.Now().Format("20060102150405") + "_" + func() string { u := utils.GenerateUID(); if len(u) > 6 { return u[:6] }; return u }()
    c.JSON(200, gin.H{
        "data": gin.H{
            "id":         id,
            "status":     "queued",
            "created_at": time.Now().Format(time.RFC3339),
            "order_id":   req.OrderID,
            "target":     req.Target,
            "priority":   req.Priority,
        },
    })
}
