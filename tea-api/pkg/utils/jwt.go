package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"tea-api/internal/config"
)

// Claims JWT声明
type Claims struct {
	UserID uint   `json:"user_id"`
	OpenID string `json:"open_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func GenerateToken(userID uint, openID string, role string) (string, error) {
	cfg := config.Config.JWT

	claims := Claims{
		UserID: userID,
		OpenID: openID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.ExpiresTime) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ParseToken 解析JWT token
func ParseToken(tokenString string) (*Claims, error) {
	cfg := config.Config.JWT

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
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

// RefreshToken 刷新token
func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查是否在刷新期内
	cfg := config.Config.JWT
	if time.Until(claims.ExpiresAt.Time) > time.Duration(cfg.BufferTime)*time.Second {
		return tokenString, nil
	}

	// 生成新token
	return GenerateToken(claims.UserID, claims.OpenID, claims.Role)
}
