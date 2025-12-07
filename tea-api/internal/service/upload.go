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
