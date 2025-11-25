package env

import "os"

// Get returns environment variable value or default if not set
func Get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
