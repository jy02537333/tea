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
