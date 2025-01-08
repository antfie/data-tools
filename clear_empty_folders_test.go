package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"path/filepath"

	"testing"
)

func TestClearEmptyFolder(t *testing.T) {
	tempTestDataPath := createTempTestDataPath(t)
	defer os.RemoveAll(tempTestDataPath)

	testFile1 := path.Join(tempTestDataPath, "/a/b/c/d/e/f/test.txt")
	err := os.MkdirAll(filepath.Dir(testFile1), 0750)
	assert.NoError(t, err)

	testFile2 := path.Join(tempTestDataPath, "/j/v/c/d/test.txt")
	err = os.MkdirAll(filepath.Dir(testFile2), 0750)
	assert.NoError(t, err)

	err = os.WriteFile(testFile1, nil, 0600)
	assert.NoError(t, err)

	err = os.WriteFile(testFile2, nil, 0600)
	assert.NoError(t, err)

	err = os.Remove(testFile1)
	assert.NoError(t, err)

	err = os.Remove(testFile2)
	assert.NoError(t, err)

	err = clearEmptyFolder(tempTestDataPath)
	assert.NoError(t, err)

	folderCount, fileCount := getFolderAndFileTotalCount(t, tempTestDataPath)

	assert.Equal(t, 4, folderCount)
	assert.Equal(t, 4, fileCount)
}
