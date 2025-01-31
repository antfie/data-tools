package main

import (
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"
)

var testDataPath = path.Join("test", "data")

func createEmptyTempTestDataPath(t *testing.T) string {
	tempTestDataPath, err := os.MkdirTemp("", "data-tools-")
	assert.NoError(t, err)

	tempTestDataAbsolutePath, err := filepath.Abs(tempTestDataPath)
	assert.NoError(t, err)

	return tempTestDataAbsolutePath
}

func createTempTestDataPath(t *testing.T) string {
	testingDataDestinationPath := createEmptyTempTestDataPath(t)

	testingDataSourcePath, err := filepath.Abs(testDataPath)
	assert.NoError(t, err)

	// Populate test data
	err = CopyOrMoveFiles(testingDataSourcePath, testingDataDestinationPath, false, false)
	assert.NoError(t, err)

	return testingDataDestinationPath
}

func getFolderAndFileTotalCount(t *testing.T, path string) (int, int) {
	folderCount := 0
	fileCount := 0

	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Ignore the root directory
		if p == path {
			return nil
		}

		if d.IsDir() {
			folderCount++
		} else {
			fileCount++
		}

		return nil
	})

	assert.NoError(t, err)
	return folderCount, fileCount
}

func (ctx *Context) AssertDBCount(t *testing.T, query string, expectedCount int) {
	var countResult int
	result := ctx.DB.Raw(query).First(&countResult)

	assert.NoError(t, result.Error)
	assert.Equal(t, int64(1), result.RowsAffected)
	assert.Equal(t, expectedCount, countResult)
}
