//go:build sqlite
// +build sqlite

package database

import (
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// initSQLite opens a SQLite database file with GORM (compiled only when build tag 'sqlite' is provided)
func initSQLite(path string) (*gorm.DB, error) {
	if path == "" {
		path = "tea_dev_api.db"
	}
	// Ensure directory exists if path contains folders
	if err := os.MkdirAll(".", 0755); err != nil {
		log.Printf("create data dir err: %v", err)
	}
	return gorm.Open(sqlite.Open(path), &gorm.Config{})
}

// SqliteCompiled indicates whether sqlite support was compiled into the binary.
func SqliteCompiled() bool { return true }
