package main

import (
	"data-tools/config"
	"gorm.io/gorm"
)

type Context struct {
	Config *config.Config
	DB     *gorm.DB
}
