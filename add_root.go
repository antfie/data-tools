package main

import (
	"data-tools-2025/models"
	"data-tools-2025/utils"
	"errors"
	"gorm.io/gorm"
	"path/filepath"
)

func (ctx *Context) AddRootPath(rootPath string) error {
	if !IsDir(rootPath) {
		return ErrCouldNotResolvePath
	}

	absoluteRootPath, err := filepath.Abs(rootPath)

	if err != nil {
		return ErrCouldNotResolvePath
	}

	utils.ConsoleAndLogPrintf("Adding root path \"%s\"", absoluteRootPath)

	var pathModel = models.Path{
		Name:  absoluteRootPath,
		Level: 0,
	}
	result := ctx.DB.First(&pathModel)

	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	// Has the path already been added?
	if result.RowsAffected > 0 {
		return ErrPathAlreadyAdded
	}

	// Add the root path
	result = ctx.DB.Create(&models.Root{
		Path: pathModel,
	})

	return result.Error
}
