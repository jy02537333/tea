package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"tea-api/internal/service"
	"tea-api/pkg/utils"
)

type UploadHandler struct {
	uploadService *service.UploadService
}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{uploadService: service.NewUploadService()}
}

// UploadMedia 处理管理端文件上传请求
func (h *UploadHandler) UploadMedia(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.InvalidParam(c, "请选择要上传的文件")
		return
	}
	defer file.Close()

	url, err := h.uploadService.UploadImage(file, header.Filename, header.Size)
	if err != nil {
		utils.Error(c, utils.CodeError, "上传失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"url": url})
}

// GetOSSPolicy 管理端获取 OSS 直传策略（用于前端直传）
// GET /api/v1/admin/storage/oss/policy?dir=admin/products/2025/12/20/&expire=1800
func (h *UploadHandler) GetOSSPolicy(c *gin.Context) {
	dir := c.Query("dir")
	exp := 0
	if v := c.Query("expire"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			exp = n
		}
	}
	data, err := h.uploadService.GenerateOSSPolicy(exp, dir)
	if err != nil {
		utils.Error(c, utils.CodeError, "获取策略失败: "+err.Error())
		return
	}
	utils.Success(c, data)
}
