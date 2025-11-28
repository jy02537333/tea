package service

import (
	"fmt"
	"os"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OssService struct {
	client *oss.Client
	bucket string
}

func NewOssService() *OssService {
	endpoint := os.Getenv("OSS_ENDPOINT")
	accessKey := os.Getenv("OSS_ACCESS_KEY_ID")
	secretKey := os.Getenv("OSS_ACCESS_KEY_SECRET")
	bucket := os.Getenv("OSS_BUCKET")
	client, err := oss.New(endpoint, accessKey, secretKey)
	if err != nil {
		panic(fmt.Sprintf("OSS初始化失败: %v", err))
	}
	return &OssService{client: client, bucket: bucket}
}

// DeleteFilesFunc is a package-level hook to allow tests to override OSS deletion behavior.
// By default it uses a real OssService instance.
var DeleteFilesFunc = func(urls []string) error {
	return NewOssService().DeleteFiles(urls)
}

// ListFiles 列出 OSS 文件（支持前缀、分页）
func (s *OssService) ListFiles(prefix string, marker string, limit int) ([]string, string, error) {
	bkt, err := s.client.Bucket(s.bucket)
	if err != nil {
		return nil, "", err
	}
	opt := []oss.Option{
		oss.Prefix(prefix),
		oss.MaxKeys(limit),
	}
	if marker != "" {
		opt = append(opt, oss.Marker(marker))
	}
	res, err := bkt.ListObjects(opt...)
	if err != nil {
		return nil, "", err
	}
	var urls []string
	// 生成对外可访问 URL（基于 bucket + endpoint + key）
	endpoint := s.client.Config.Endpoint
	for _, obj := range res.Objects {
		urls = append(urls, "https://"+s.bucket+"."+endpoint+"/"+obj.Key)
	}
	nextMarker := res.NextMarker
	return urls, nextMarker, nil
}

// DeleteFiles 批量删除 OSS 文件
func (s *OssService) DeleteFiles(urls []string) error {
	bkt, err := s.client.Bucket(s.bucket)
	if err != nil {
		return err
	}
	var keys []string
	for _, url := range urls {
		// 只取 key 部分，寻找 bucket 名称在 URL 中的位置
		idx := strings.Index(url, s.bucket)
		if idx >= 0 {
			key := url[idx+len(s.bucket)+1:]
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return nil
	}
	_, err = bkt.DeleteObjects(keys)
	return err
}
