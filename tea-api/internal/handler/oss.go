package handler

import (
	"fmt"
	"net/http"

	"tea-api/internal/service"
	"tea-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type OssHandler struct {
	ossService *service.OssService
}

func NewOssHandler(ossService *service.OssService) *OssHandler {
	return &OssHandler{ossService: ossService}
}

// ListOssFiles 列出 OSS 文件（支持前缀、分页）
func (h *OssHandler) ListOssFiles(c *gin.Context) {
	prefix := c.Query("prefix")
	marker := c.Query("marker")
	limit := 100
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	files, nextMarker, err := h.ossService.ListFiles(prefix, marker, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "OSS图片列表获取失败", err.Error())
		return
	}
	c.JSON(200, gin.H{
		"files":       files,
		"next_marker": nextMarker,
	})
}

// DeleteOssFile 删除 OSS 文件（支持批量）
func (h *OssHandler) DeleteOssFile(c *gin.Context) {
	var req struct {
		Urls []string `json:"urls" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Urls) == 0 {
		response.Error(c, http.StatusBadRequest, "参数错误", "urls 不能为空")
		return
	}
	if err := h.ossService.DeleteFiles(req.Urls); err != nil {
		response.Error(c, http.StatusInternalServerError, "OSS图片删除失败", err.Error())
		return
	}
	response.Success(c, nil)
}
