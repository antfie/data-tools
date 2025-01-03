package main

import (
	"github.com/stretchr/testify/assert"
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
	tempTestDataPath := createEmptyTempTestDataPath(t)

	testDataAbsolutePath, err := filepath.Abs(testDataPath)
	assert.NoError(t, err)

	// Populate test data
	err = CopyOrMoveFiles(testDataAbsolutePath, tempTestDataPath, false)
	assert.NoError(t, err)

	return tempTestDataPath
}
