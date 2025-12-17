package utils

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	// 统一改用新版 JWT 配置来源
	"tea-api/internal/pkg/jwtcfg"
)

// Claims JWT声明（保持旧结构以兼容调用方）
type Claims struct {
	UserID uint   `json:"user_id"`
	OpenID string `json:"open_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token（使用新版 jwtcfg 配置）
func GenerateToken(userID uint, openID string, role string) (string, error) {
	cfg := jwtcfg.Get()

	claims := Claims{
		UserID: userID,
		OpenID: openID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    getIssuer(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.ExpiryMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ParseToken 解析JWT token（使用新版 jwtcfg 配置）
func ParseToken(tokenString string) (*Claims, error) {
	cfg := jwtcfg.Get()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// RefreshToken 刷新token（使用新版 jwtcfg 配置）
func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}
	// 若未临近过期（缓冲窗口内不刷新），直接返回原token
	buffer := getBufferMinutes()
	if time.Until(claims.ExpiresAt.Time) > time.Duration(buffer)*time.Minute {
		return tokenString, nil
	}
	return GenerateToken(claims.UserID, claims.OpenID, claims.Role)
}

func getIssuer() string {
	if v := os.Getenv("TEA_JWT_ISSUER"); v != "" {
		return v
	}
	return "tea-api"
}

func getBufferMinutes() int {
	if v := os.Getenv("TEA_JWT_BUFFER_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return 10
}
