package main

import (
	"flag"
	"fmt"
	"log"
	"tea-api/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// gen_admin_token generates a temporary admin JWT; issuer forced to "tea-api" for unified runtime
func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "tea-api/configs/config.yaml", "path to tea-api config.yaml")
	flag.Parse()

	if err := config.LoadConfig(cfgPath); err != nil {
		log.Fatalf("load config: %v", err)
	}

	secret := config.Cfg.JWT.Secret
	claims := jwt.MapClaims{
		"user_id": 1,
		"open_id": "admin_openid_temp",
		"role":    "admin",
		"iss":     "tea-api",
		"iat":     time.Now().Unix(),
		"nbf":     time.Now().Unix(),
		"exp":     time.Now().Add(time.Duration(config.Cfg.JWT.ExpiryMinutes) * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Fatalf("sign token: %v", err)
	}
	fmt.Println(s)
}
