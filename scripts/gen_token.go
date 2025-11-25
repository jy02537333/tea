package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// NOTE: keep this in sync with configs/config.yaml jwt.secret for testing only
	secret := "tea-shop-jwt-secret-key-2023"
	// claims similar to `utils.Claims` in repo
	claims := jwt.MapClaims{
		"user_id": 1,
		"open_id": "admin_openid",
		"role":    "admin",
		"iss":     "tea-shop",
		"iat":     time.Now().Unix(),
		"nbf":     time.Now().Unix(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
}
