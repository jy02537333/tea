//go:build !sqlite
// +build !sqlite

package database

import (
	"errors"

	"gorm.io/gorm"
)

// initSQLite stub when sqlite build tag is not provided.
func initSQLite(path string) (*gorm.DB, error) {
	return nil, errors.New("sqlite support not compiled (build with -tags=sqlite)")
}

// SqliteCompiled indicates whether sqlite support was compiled into the binary.
func SqliteCompiled() bool { return false }
