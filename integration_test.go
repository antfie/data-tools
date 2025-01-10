package main

import (
	"data-tools/config"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestIntegration(t *testing.T) {
	tempTestDataPath := createTempTestDataPath(t)
	zapTestDataPath := createEmptyTempTestDataPath(t)
	defer os.RemoveAll(tempTestDataPath)
	defer os.RemoveAll(zapTestDataPath)

	c := &config.Config{
		DBPath:                      path.Join(zapTestDataPath, "db.db"),
		BatchSize:                   5,
		MaxConcurrentFileOperations: 2,
	}

	ctx := &Context{
		Config: c,
		DB:     initDb(c),
	}

	err := ctx.AddRootPath(tempTestDataPath)
	assert.NoError(t, err)

	err = ctx.Crawl()
	assert.NoError(t, err)

	err = ctx.HashFiles()
	assert.NoError(t, err)

	err = ctx.TypeFiles()
	assert.NoError(t, err)

	err = ctx.SizeFiles()
	assert.NoError(t, err)

	err = ctx.Zap(zapTestDataPath, false)
	assert.NoError(t, err)

	folderCount, fileCount := getFolderAndFileTotalCount(t, tempTestDataPath)
	assert.Zero(t, folderCount)
	assert.Zero(t, fileCount)
}
