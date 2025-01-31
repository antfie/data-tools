//go:build integration
// +build integration

package main

import (
	"data-tools/config"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestZapFileIntegrity(t *testing.T) {
	tempTestDataPath := createTempTestDataPath(t)
	defer os.RemoveAll(tempTestDataPath)

	zapDatapath := path.Join(tempTestDataPath, "ZAP")

	c := &config.Config{
		DBPath:                      path.Join(tempTestDataPath, "db.db"),
		BatchSize:                   5,
		MaxConcurrentFileOperations: 2,
		ZapDataPath:                 zapDatapath,
		IsDebug:                     true,
	}

	ctx := &Context{
		Config: c,
		DB:     initDb(c),
	}

	dataPath := path.Join(tempTestDataPath, "a")
	err := ctx.Crawl(dataPath)
	assert.NoError(t, err)

	err = ctx.HashFiles()
	assert.NoError(t, err)

	err = ctx.Zap(false)
	assert.NoError(t, err)

	folderCount, fileCount := getFolderAndFileTotalCount(t, dataPath)
	assert.Zero(t, folderCount)
	assert.Zero(t, fileCount)

	err = ctx.ZapDBIntegrityTestBySize()
	assert.NoError(t, err)

	ctx.AssertDBCount(t, "SELECT COUNT(*) FROM file_hashes WHERE zapped = 1", 3)

	// Corrupt a file
	zappedFilePath := path.Join(zapDatapath, "4f/57/8179952b85b92c2b464c64fabc6134fa0fa9692c8333cbe1c48cf6eeb9bc89b4f91338681a12f377b6cda17643ae3b4a18849f99f20ab7b7873dc95b3355")
	assert.True(t, IsFile(zappedFilePath))

	err = os.WriteFile(zappedFilePath, []byte("h"), 0600)
	assert.NoError(t, err)

	// Corrupt another file
	zappedFilePath = path.Join(zapDatapath, "90/65/133a01270fbc15e2428f4b6318d4c6b0ef85803c272aeb64ce416e6e51df4cecdc6ca95b888781f875ea112bd73f88e1c187ca17254bf911fd70216800b0")
	assert.True(t, IsFile(zappedFilePath))

	err = os.WriteFile(zappedFilePath, []byte("h"), 0600)
	assert.NoError(t, err)

	err = ctx.ZapDBIntegrityTestBySize()
	assert.NoError(t, err)

	ctx.AssertDBCount(t, "SELECT COUNT(*) FROM file_hashes WHERE zapped = 1", 1)
}
