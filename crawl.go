package main

import (
	"data-tools/models"
	"data-tools/utils"
	"errors"
	"gorm.io/gorm"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func (ctx *Context) Crawl(rootPath string) error {
	absoluteRootPath, err := filepath.Abs(rootPath)

	if err != nil {
		return ErrCouldNotResolvePath
	}

	if !IsDir(absoluteRootPath) {
		return ErrCouldNotResolvePath
	}

	var pathModel models.Path
	result := ctx.DB.Where("name = ?", absoluteRootPath).First(&pathModel)

	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	// Has the path already been added, hence crawled?
	if result.RowsAffected > 0 {
		utils.ConsoleAndLogPrintf("\"%s\" has already been crawled.", absoluteRootPath)
		return ErrPathAlreadyAdded
	}

	rootPathModel := models.Path{
		Name: absoluteRootPath,
	}

	// Add the root path
	result = ctx.DB.Create(&rootPathModel)

	if result.Error != nil {
		return result.Error
	}

	utils.ConsoleAndLogPrintf("Crawling \"%s\"", absoluteRootPath)
	return ctx.crawlRootPath(rootPathModel)
}

func (ctx *Context) crawlRootPath(rootPath models.Path) error {
	rootPathSeparatorCount := getPathSeparatorCount(rootPath.Name)
	currentLevel := uint(0)
	pathModels := map[string]*models.Path{rootPath.Name: &rootPath}
	pathCount := int64(0)
	fileCount := int64(0)

	err := ctx.DB.Transaction(func(tx *gorm.DB) error {
		return filepath.WalkDir(rootPath.Name, func(thisPath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			levelCalculationAsInt := getPathSeparatorCount(thisPath) - rootPathSeparatorCount

			// Ensure we have not wrapped around for uint conversion, prevent CWE-190
			if levelCalculationAsInt < 0 {
				return nil
			}

			currentLevel = uint(levelCalculationAsInt)

			if d.IsDir() {
				// Ignore level 0 directories
				if currentLevel == 0 {
					return nil
				}

				if utils.IsInArray(d.Name(), ctx.Config.FolderNamesToIgnore) {
					return filepath.SkipDir
				}

				pathModels[thisPath] = &models.Path{
					ParentPath: pathModels[filepath.Dir(thisPath)],
					Name:       d.Name(),
					Level:      currentLevel,
				}

				result := tx.Create(pathModels[thisPath])

				if result.Error != nil {
					return result.Error
				}

				pathCount++
			} else {
				if utils.IsInArray(d.Name(), ctx.Config.FileNamesToIgnore) {
					return nil
				}

				result := tx.Create(&models.File{
					Path:  *pathModels[filepath.Dir(thisPath)],
					Name:  d.Name(),
					Level: currentLevel,
				})

				if result.Error != nil {
					return result.Error
				}

				fileCount++
			}

			return nil
		})
	})

	// Output a summary
	if err == nil {
		utils.ConsoleAndLogPrintf("Found %s and %s", utils.Pluralize("path", pathCount), utils.Pluralize("file", fileCount))
	}

	return err
}

func getPathSeparatorCount(path string) int {
	return strings.Count(path, string(os.PathSeparator))
}
