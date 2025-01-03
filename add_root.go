package main

import (
	"data-tools/models"
	"data-tools/utils"
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

	var pathModel models.Path
	result := ctx.DB.Where("name = ?", absoluteRootPath).First(&pathModel)

	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	// Has the path already been added?
	if result.RowsAffected > 0 {
		utils.ConsoleAndLogPrintf("Root path \"%s\" has already been added.", absoluteRootPath)
		return ErrPathAlreadyAdded
	}

	// Add the root path
	result = ctx.DB.Create(&models.Root{
		Path: pathModel,
	})

	return result.Error
}
