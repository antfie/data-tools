//go:build alternative_driver

package main

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func GetDriver(dsn string, gormConfig *gorm.Config) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dsn), gormConfig)
}
