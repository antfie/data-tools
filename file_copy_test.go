package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestFileCopyShouldCreateDirectoryAndCopy(t *testing.T) {
	tempTestDataPath := createTempTestDataPath(t)
	defer os.RemoveAll(tempTestDataPath)

	sourceFilePath := path.Join(tempTestDataPath, "/a/file.md")
	destinationFilePath := path.Join(tempTestDataPath, "/v/file2.txt")

	success, err := CopyOrMoveFile(sourceFilePath, destinationFilePath, false, false)
	assert.NoError(t, err)
	assert.True(t, success)

	filesEqual, err := CompareFiles(sourceFilePath, destinationFilePath)
	assert.NoError(t, err)
	assert.True(t, filesEqual)
}

func TestFileCopyShouldErrorIfDirectoryDoesNotExistWhenZapIsSet(t *testing.T) {
	tempTestDataPath := createTempTestDataPath(t)
	defer os.RemoveAll(tempTestDataPath)

	sourceFilePath := path.Join(tempTestDataPath, "/a/file.md")
	destinationFilePath := path.Join(tempTestDataPath, "/v/file2.txt")

	success, err := CopyOrMoveFile(sourceFilePath, destinationFilePath, true, true)
	assert.Error(t, err)
	assert.False(t, success)

	assert.True(t, IsFile(sourceFilePath))
	assert.False(t, IsFile(destinationFilePath))
}

func TestFileCopyShouldDoNothingIfTheFileIsDifferent(t *testing.T) {
	tempTestDataPath := createTempTestDataPath(t)
	defer os.RemoveAll(tempTestDataPath)

	sourceFilePath := path.Join(tempTestDataPath, "/a/file.md")
	destinationFilePath := path.Join(tempTestDataPath, "/a/b/4276652.png")

	success, err := CopyOrMoveFile(sourceFilePath, destinationFilePath, true, true)
	assert.NoError(t, err)
	assert.False(t, success)

	assert.True(t, IsFile(sourceFilePath))
	assert.True(t, IsFile(destinationFilePath))

	filesEqual, err := CompareFiles(sourceFilePath, destinationFilePath)
	assert.NoError(t, err)
	assert.False(t, filesEqual)
}
