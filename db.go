package main

import (
	"data-tools/config"
	"data-tools/models"
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

	return connect("file::memory:", gormConfig)
}

func connect(dsn string, gormConfig *gorm.Config) *gorm.DB {
	db, err := GetDriver(dsn, gormConfig)

	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}

	// From: https://github.com/rails/rails/blob/8c7e39497f069e354d67ed14c63fa31383871e5d/activerecord/lib/active_record/connection_adapters/sqlite3_adapter.rb#L107
	if res := db.Exec("PRAGMA foreign_keys = ON;PRAGMA journal_mode = WAL;PRAGMA synchronous = NORMAL;PRAGMA mmap_size = 134217728;PRAGMA journal_size_limit = 67108864;PRAGMA cache_size = 2000", nil); res.Error != nil {
		log.Fatalf("failed to configure the database: %v", res.Error)
	}

	err = db.AutoMigrate(
		&models.PathHash{},
		&models.Path{},
		&models.FileType{},
		&models.FileHash{},
		&models.File{},
		&models.Note{},
		&models.PathHashNote{},
		&models.PathNote{},
		&models.FileTypeNote{},
		&models.FileHashNote{},
		&models.FileNote{},
	)

	if err != nil {
		log.Fatalf("failed to migrate the database")
	}

	return db
}
