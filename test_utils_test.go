package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestF(t *testing.T) {
	folderCount, fileCount := getFolderAndFileTotalCount(t, testDataPath)

	assert.Equal(t, 4, folderCount)
	assert.Equal(t, 5, fileCount)
}
