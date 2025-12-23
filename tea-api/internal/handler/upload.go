package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"

	"tea-api/internal/service"
	"tea-api/pkg/utils"
	"tea-api/internal/config"
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

// AdminGetOSSPolicy 生成阿里云 OSS 直传所需的 policy 与签名（管理端）
// 请求示例：{"business":"product_image","file_name":"a.jpg","content_type":"image/jpeg","file_size": 12345}
// 响应字段：policy, signature, accessKeyId, expire_at, object_key_template
func (h *UploadHandler) AdminGetOSSPolicy(c *gin.Context) {
	type reqBody struct {
		Business    string `json:"business"`
		FileName    string `json:"file_name"`
		ContentType string `json:"content_type"`
		FileSize    int64  `json:"file_size"`
	}
	var req reqBody
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.InvalidParam(c, "invalid request body")
		return
	}

	// 基础配置校验
	ossCfg := config.Config.Upload.OSS
	if ossCfg.Endpoint == "" || ossCfg.AccessKeyID == "" || ossCfg.AccessKeySecret == "" || ossCfg.BucketName == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "not_ready",
			"message": "OSS 配置缺失或未就绪，请先在 configs 中配置 upload.oss",
		})
		return
	}

	// 业务前缀与对象 key 模板（按天分目录 + 原文件名），可按需扩展策略
	if req.Business == "" {
		req.Business = "uploads"
	}
	datePrefix := time.Now().UTC().Format("2006/01/02")
	keyPrefix := fmt.Sprintf("%s/%s/", req.Business, datePrefix)
	objectKeyTmpl := keyPrefix + "${filename}"

	// 过期时间（短期有效）
	expireAt := time.Now().UTC().Add(5 * time.Minute).Format("2006-01-02T15:04:05Z")

	// policy 条件（最小集：bucket / key 前缀 + 可选大小与类型范围）
	conditions := make([]interface{}, 0, 4)
	conditions = append(conditions, map[string]string{"bucket": ossCfg.BucketName})
	conditions = append(conditions, []interface{}{"starts-with", "$key", keyPrefix})
	if req.ContentType != "" {
		// 仅校验前缀，允许更具体的子类型
		conditions = append(conditions, []interface{}{"starts-with", "$Content-Type", req.ContentType})
	}
	if req.FileSize > 0 {
		conditions = append(conditions, []interface{}{"content-length-range", 0, req.FileSize})
	}

	policyDoc := map[string]interface{}{
		"expiration": expireAt,
		"conditions": conditions,
	}
	raw, err := json.Marshal(policyDoc)
	if err != nil {
		utils.Error(c, utils.CodeError, "failed to marshal policy: "+err.Error())
		return
	}

	policyBase64 := utils.Base64Encode(raw)
	signature := utils.HMACSHA1Base64(ossCfg.AccessKeySecret, policyBase64)

	c.JSON(http.StatusOK, gin.H{
		"policy":               policyBase64,
		"signature":            signature,
		"accessKeyId":          ossCfg.AccessKeyID,
		"expire_at":            expireAt,
		"object_key_template":  objectKeyTmpl,
	})
}
