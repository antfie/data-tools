//go:build !alternative_driver

package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// GetDriver requires CGO for this implementation
func GetDriver(dsn string, gormConfig *gorm.Config) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dsn), gormConfig)
}
