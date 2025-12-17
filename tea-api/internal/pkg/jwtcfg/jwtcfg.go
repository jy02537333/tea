package jwtcfg

import (
	"os"
	"strconv"
)

type Config struct {
	Secret        string
	ExpiryMinutes int
}

// Get returns JWT config from env vars with sane defaults.
// TEA_JWT_SECRET: secret string
// TEA_JWT_EXP_MIN: expiry minutes (int)
func Get() Config {
	secret := os.Getenv("TEA_JWT_SECRET")
	if secret == "" {
		secret = "dev_secret_change_me"
	}

	expMin := 120
	if v := os.Getenv("TEA_JWT_EXP_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			expMin = n
		}
	}
	return Config{Secret: secret, ExpiryMinutes: expMin}
}
