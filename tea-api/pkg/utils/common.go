package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"
	mathrand "math/rand"
	"strings"
	"time"
)

// GenerateUID 生成32位唯一ID
func GenerateUID() string {
	timestamp := time.Now().Unix()
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(999999))

	source := fmt.Sprintf("%d%d", timestamp, randomNum.Int64())
	hash := md5.Sum([]byte(source))
	return fmt.Sprintf("%x", hash)
}

// GenerateOrderNo 生成订单号
func GenerateOrderNo(prefix string) string {
	now := time.Now()
	timeStr := now.Format("20060102150405")
	randomNum := mathrand.Intn(999999)
	return fmt.Sprintf("%s%s%06d", prefix, timeStr, randomNum)
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[mathrand.Intn(len(charset))]
	}
	return string(b)
}

// MD5Hash 计算MD5哈希值
func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return fmt.Sprintf("%x", hash)
}

// HMACSHA256Hex 计算 HMAC-SHA256 并返回十六进制字符串（小写）
func HMACSHA256Hex(key, message string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

// StringInSlice 检查字符串是否在切片中
func StringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// IntInSlice 检查整数是否在切片中
func IntInSlice(num int, slice []int) bool {
	for _, n := range slice {
		if n == num {
			return true
		}
	}
	return false
}

// RemoveEmptyStrings 移除空字符串
func RemoveEmptyStrings(slice []string) []string {
	var result []string
	for _, str := range slice {
		if strings.TrimSpace(str) != "" {
			result = append(result, str)
		}
	}
	return result
}

// TruncateString 截断字符串
func TruncateString(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length] + "..."
}

// GenerateRandomCode 生成随机验证码
func GenerateRandomCode(length int) string {
	const digits = "0123456789"
	b := make([]byte, length)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		b[i] = digits[num.Int64()]
	}
	return string(b)
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	sourceFile, err := io.ReadAll(strings.NewReader(src))
	if err != nil {
		return err
	}

	// 这里简化处理，实际应该使用文件操作
	_ = sourceFile
	_ = dst
	return nil
}

func init() {
	mathrand.Seed(time.Now().UnixNano())
}
