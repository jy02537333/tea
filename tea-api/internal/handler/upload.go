package handler

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"tea-api/internal/config"
	"tea-api/internal/service"
	"tea-api/pkg/response"
)

type UploadHandler struct{ uploadService *service.UploadService }

func NewUploadHandler() *UploadHandler { return &UploadHandler{uploadService: service.NewUploadService()} }

// OSSPolicy 颁发阿里云 OSS 直传策略
// POST /api/v1/upload/oss/policy  {"dir":"uploads/20260109/"}
// 返回 {accessid, host, policy, signature, dir, expire}
func (h *UploadHandler) OSSPolicy(c *gin.Context) {
    var req struct{ Dir string `json:"dir"` }
    _ = c.ShouldBindJSON(&req)

    cfg := config.Config.Upload.OSS
    if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" || cfg.BucketName == "" {
        response.Error(c, http.StatusBadRequest, "OSS 未配置")
        return
    }

    // 默认目录：uploads/YYYYMMDD/
    if req.Dir == "" {
        req.Dir = "uploads/" + time.Now().Format("20060102") + "/"
    }

    expireSeconds := int64(60) // 1 分钟有效期
    expireAt := time.Now().Add(time.Duration(expireSeconds) * time.Second)
    expiration := expireAt.UTC().Format("2006-01-02T15:04:05Z")

    // 构造 policy 条件：key 以 dir 开头，限制大小 0~10MB
    cond := []interface{}{
        map[string]string{"bucket": cfg.BucketName},
        []interface{}{"starts-with", "$key", req.Dir},
        []interface{}{"content-length-range", 0, 10 * 1024 * 1024},
    }
    policyDoc := map[string]interface{}{
        "expiration": expiration,
        "conditions": cond,
    }
    bs, _ := json.Marshal(policyDoc)
    policy := base64.StdEncoding.EncodeToString(bs)

    mac := hmac.New(sha1.New, []byte(cfg.AccessKeySecret))
    _, _ = mac.Write([]byte(policy))
    sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

    host := "https://" + cfg.BucketName + "." + cfg.Endpoint

    resp := map[string]interface{}{
        "accessid": cfg.AccessKeyID,
        "host":     host,
        "policy":   policy,
        "signature": sign,
        "dir":      req.Dir,
        "expire":   expireAt.Unix(),
    }
    response.Success(c, resp)
}

// UploadMedia 传统表单上传（管理端复用）：接收 multipart 文件后转存 OSS，返回 URL
func (h *UploadHandler) UploadMedia(c *gin.Context) {
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        response.BadRequest(c, "请选择要上传的文件")
        return
    }
    defer file.Close()

    url, err := h.uploadService.UploadImage(file, header.Filename, header.Size)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "上传失败: "+err.Error())
        return
    }
    response.Success(c, gin.H{"url": url})
}
