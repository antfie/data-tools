package main

import (
	"data-tools/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAllFilesRelativeToRootPath(t *testing.T) {
	ctx := &Context{
		Config: &config.Config{},
		DB:     testDB(),
	}

	err := ctx.AddRootPath(testDataPath)
	assert.NoError(t, err)

	err = ctx.Crawl()
	assert.NoError(t, err)
}
