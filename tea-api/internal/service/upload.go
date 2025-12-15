package service

import (
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

// OSSPolicyResponse OSS上传策略响应
type OSSPolicyResponse struct {
	AccessKeyID     string `json:"access_key_id"`
	Policy          string `json:"policy"`
	Signature       string `json:"signature"`
	Expire          int64  `json:"expire"`
	Host            string `json:"host"`
	ObjectKeyPrefix string `json:"object_key_prefix"`
}

// GenerateOSSPolicy 生成OSS直传策略
func (s *UploadService) GenerateOSSPolicy(business string) (*OSSPolicyResponse, error) {
	_, err := s.ensureBucket()
	if err != nil {
		return nil, err
	}

	cfg := config.Config.Upload.OSS
	
	// 设置过期时间（30分钟）
	expireTime := time.Now().Add(30 * time.Minute)
	expireTimestamp := expireTime.Unix()

	// 构建对象key前缀
	datePath := time.Now().Format("2006/01/02")
	var objectKeyPrefix string
	switch business {
	case "product_image":
		objectKeyPrefix = fmt.Sprintf("admin/products/%s/", datePath)
	case "brand_logo":
		objectKeyPrefix = fmt.Sprintf("admin/brands/%s/", datePath)
	case "store_image":
		objectKeyPrefix = fmt.Sprintf("admin/stores/%s/", datePath)
	default:
		objectKeyPrefix = fmt.Sprintf("admin/uploads/%s/", datePath)
	}

	// 构建host
	endpoint := cfg.Endpoint
	scheme := "https"
	if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.TrimPrefix(endpoint, "https://")
	} else if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
		scheme = "http"
	}
	host := fmt.Sprintf("%s://%s.%s", scheme, cfg.BucketName, endpoint)

	// 生成policy和signature
	policy, signature := generatePolicyAndSignature(cfg.AccessKeyID, cfg.AccessKeySecret, objectKeyPrefix, expireTimestamp)

	return &OSSPolicyResponse{
		AccessKeyID:     cfg.AccessKeyID,
		Policy:          policy,
		Signature:       signature,
		Expire:          expireTimestamp,
		Host:            host,
		ObjectKeyPrefix: objectKeyPrefix,
	}, nil
}

// generatePolicyAndSignature 生成OSS policy和签名
func generatePolicyAndSignature(accessKeyID, accessKeySecret, dir string, expireTimestamp int64) (string, string) {
	// 简化实现：生成基础policy
	policy := fmt.Sprintf(`{"expiration":"%s","conditions":[["content-length-range",0,10485760],["starts-with","$key","%s"]]}`,
		time.Unix(expireTimestamp, 0).UTC().Format("2006-01-02T15:04:05.000Z"),
		dir,
	)
	
	// Base64编码policy
	policyBase64 := utils.Base64Encode([]byte(policy))
	
	// 使用HMAC-SHA1签名
	signature := utils.HmacSha1Sign(accessKeySecret, policyBase64)
	
	return policyBase64, signature
}
