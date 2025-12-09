package env

import "os"

// Get returns the value of key if it is set, otherwise returns defaultValue.
func Get(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// MustGet returns the value of key or panics if the variable is not set.
func MustGet(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	panic("environment variable not set: " + key)
}
