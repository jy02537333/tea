package handler

import (
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

// GetOSSPolicy 获取OSS直传策略
// POST /api/v1/admin/storage/oss/policy
func (h *UploadHandler) GetOSSPolicy(c *gin.Context) {
	type PolicyRequest struct {
		Business string `json:"business" binding:"required"` // product_image/brand_logo/store_image
	}

	var req PolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.InvalidParam(c, err.Error())
		return
	}

	policy, err := h.uploadService.GenerateOSSPolicy(req.Business)
	if err != nil {
		utils.Error(c, utils.CodeError, "生成上传策略失败: "+err.Error())
		return
	}

	utils.Success(c, policy)
}
