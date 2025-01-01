package main

import (
	"data-tools/config"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func TestAddRootGivenAFilePassedShouldReturnError(t *testing.T) {
	ctx := &Context{
		Config: &config.Config{},
	}

	err := ctx.AddRootPath(path.Join(testDataPath, "a/b/j.txt"))

	assert.ErrorIs(t, err, ErrCouldNotResolvePath)
}

func TestAddRootGivenAnInvalidPathReturnError(t *testing.T) {
	ctx := &Context{
		Config: &config.Config{},
	}

	err := ctx.AddRootPath(path.Join(testDataPath, "fail/"))

	assert.ErrorIs(t, err, ErrCouldNotResolvePath)
}
