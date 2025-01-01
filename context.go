package main

import (
	"data-tools-2025/config"
	"gorm.io/gorm"
)

type Context struct {
	Config *config.Config
	DB     *gorm.DB
}
