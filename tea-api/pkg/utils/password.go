package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plaintext password using bcrypt.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPasswordHash compares a plaintext password with a bcrypt hash.
func CheckPasswordHash(hash, password string) bool {
	if hash == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
