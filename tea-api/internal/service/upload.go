package service

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"tea-api/internal/config"
	"tea-api/pkg/utils"
)

// UploadService 提供文件上传到 OSS 的能力
type UploadService struct {
	once   sync.Once
	bucket *oss.Bucket
	err    error
}

func NewUploadService() *UploadService {
	return &UploadService{}
}

// UploadImage 上传文件到 OSS，返回可公开访问的 URL
func (s *UploadService) UploadImage(reader io.Reader, filename string, size int64) (string, error) {
	bucket, err := s.ensureBucket()
	if err != nil {
		return "", err
	}

	objectKey := buildObjectKey(filename)
	options := make([]oss.Option, 0, 2)
	if ct := contentTypeFromExt(objectKey); ct != "" {
		options = append(options, oss.ContentType(ct))
	}
	if size > 0 {
		options = append(options, oss.ContentLength(size))
	}

	if err := bucket.PutObject(objectKey, reader, options...); err != nil {
		return "", err
	}

	return buildOSSURL(objectKey), nil
}

func (s *UploadService) ensureBucket() (*oss.Bucket, error) {
	s.once.Do(func() {
		cfg := config.Config.Upload.OSS
		if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" || cfg.BucketName == "" {
			s.err = errors.New("OSS 配置缺失")
			return
		}

		client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
		if err != nil {
			s.err = err
			return
		}
		bucket, err := client.Bucket(cfg.BucketName)
		if err != nil {
			s.err = err
			return
		}
		s.bucket = bucket
	})
	return s.bucket, s.err
}

func buildObjectKey(filename string) string {
	datePath := time.Now().Format("2006/01/02")
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".jpg"
	}
	return fmt.Sprintf("admin/products/%s/%s%s", datePath, utils.GenerateUID(), ext)
}

func contentTypeFromExt(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return ""
	}
	return mime.TypeByExtension(ext)
}

func buildOSSURL(objectKey string) string {
	cfg := config.Config.Upload.OSS
	endpoint := cfg.Endpoint
	scheme := "https"
	if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.TrimPrefix(endpoint, "https://")
	} else if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
		scheme = "http"
	}
	if endpoint == "" {
		return objectKey
	}
	return fmt.Sprintf("%s://%s.%s/%s", scheme, cfg.BucketName, endpoint, objectKey)
}

// GenerateOSSPolicy 生成浏览器直传 OSS 的签名策略
// 返回前端直传所需字段：host、dir、policy、signature、accessKeyId、expire
func (s *UploadService) GenerateOSSPolicy(expireSeconds int, dirPrefix string) (map[string]any, error) {
	cfg := config.Config.Upload.OSS
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" || cfg.BucketName == "" {
		return nil, errors.New("OSS 配置缺失")
	}

	if expireSeconds <= 0 {
		expireSeconds = 1800 // 默认 30 分钟
	}
	// 目录前缀，默认使用 admin/products/yyyy/mm/dd
	if strings.TrimSpace(dirPrefix) == "" {
		dirPrefix = fmt.Sprintf("admin/products/%s/", time.Now().Format("2006/01/02"))
	} else if !strings.HasSuffix(dirPrefix, "/") {
		dirPrefix = dirPrefix + "/"
	}

	// 构建 policy 文档：限定上传 key 以 dirPrefix 开头，大小限制 100MB，可按需调整
	expireAt := time.Now().Add(time.Duration(expireSeconds) * time.Second).UTC()
	policyDoc := fmt.Sprintf("{\"expiration\":\"%s\",\"conditions\":[[\"starts-with\",\"$key\",\"%s\"],[\"content-length-range\",0,104857600]]}", expireAt.Format(time.RFC3339), dirPrefix)
	policyBase64 := base64.StdEncoding.EncodeToString([]byte(policyDoc))

	// 计算签名：HMAC-SHA1(secret, policyBase64) 后 base64
	mac := hmac.New(sha1.New, []byte(cfg.AccessKeySecret))
	mac.Write([]byte(policyBase64))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// host 统一规范化
	endpoint := cfg.Endpoint
	if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.TrimPrefix(endpoint, "https://")
	} else if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
	}
	host := fmt.Sprintf("https://%s.%s", cfg.BucketName, endpoint)

	return map[string]any{
		"host":         host,
		"dir":          dirPrefix,
		"policy":       policyBase64,
		"signature":    signature,
		"accessKeyId":  cfg.AccessKeyID,
		"expire":       expireAt.Unix(),
	}, nil
}
