package main

import (
	"data-tools/models"
	"data-tools/utils"
	"github.com/schollz/progressbar/v3"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func (ctx *Context) Crawl() error {
	var rootPaths []models.Path
	result := ctx.DB.Where("child_path_count IS NULL AND deleted_at IS NULL AND level = 0 AND ignored = 0").Find(&rootPaths)

	if result.Error != nil {
		return nil
	}

	// Nothing to do
	if len(rootPaths) == 0 {
		utils.ConsoleAndLogPrintf("No paths to crawl. Have you added a root path?")
		return nil
	}

	utils.ConsoleAndLogPrintf("Crawling %s", utils.Pluralize("root path", result.RowsAffected))
	bar := progressbar.Default(result.RowsAffected)

	for _, rootPath := range rootPaths {
		err := ctx.crawlRootPath(rootPath)

		if err != nil {
			return err
		}

		err = bar.Add(1)

		if err != nil {
			log.Printf("failed to update progress bar: %v", err)
		}
	}

	return nil
}

func (ctx *Context) crawlRootPath(rootPath models.Path) error {
	rootPathSeparatorCount := getSeparatorCount(rootPath.Name)
	currentLevel := uint(0)
	px := map[string]*models.Path{rootPath.Name: &rootPath}

	return filepath.WalkDir(rootPath.Name, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		levelCalculationAsInt := getSeparatorCount(p) - rootPathSeparatorCount

		// Ensure we have not wrapped around for uint conversion
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

			px[p] = &models.Path{
				ParentPath: px[filepath.Dir(p)],
				Name:       d.Name(),
				Level:      currentLevel,
			}

			result := ctx.DB.Create(px[p])

			if result.Error != nil {
				return result.Error
			}
		} else {
			if utils.IsInArray(d.Name(), ctx.Config.FileNamesToIgnore) {
				return nil
			}

			result := ctx.DB.Create(&models.File{
				Path:  *px[filepath.Dir(p)],
				Name:  d.Name(),
				Level: currentLevel,
			})

			if result.Error != nil {
				return result.Error
			}
		}

		return nil
	})
}

func getSeparatorCount(path string) int {
	return strings.Count(path, string(os.PathSeparator))
}
