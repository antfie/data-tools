package main

import (
	"data-tools/config"
	"data-tools/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

func initDb(config *config.Config) *gorm.DB {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(getLogLevel(config)),
	}

	return connect(config.DBPath, gormConfig)
}

func getLogLevel(config *config.Config) logger.LogLevel {
	if config.IsDebug {
		return logger.Info
	}

	return logger.Silent
}

func testDB() *gorm.DB {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	return connect("file::memory:?cache=shared", gormConfig)
}

func connect(dsn string, gormConfig *gorm.Config) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dsn), gormConfig)

	if err != nil {
		log.Fatalf("failed to connect to the database")
	}

	err = db.AutoMigrate(&models.Path{}, &models.Root{}, &models.File{})

	if err != nil {
		log.Fatalf("failed to migrate the database")
	}

	return db
}
