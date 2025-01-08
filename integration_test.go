package main

import (
	"data-tools/config"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestIntegration(t *testing.T) {
	tempTestDataPath := createTempTestDataPath(t)
	zapTestDataPath := createEmptyTempTestDataPath(t)
	defer os.RemoveAll(tempTestDataPath)
	defer os.RemoveAll(zapTestDataPath)

	ctx := &Context{
		Config: &config.Config{
			BatchSize:                   5,
			MaxConcurrentFileOperations: 2,
		},
		DB: testDB(),
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
}
