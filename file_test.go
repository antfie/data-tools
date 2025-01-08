package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var pathForInputs = []string{
	"/j/c/t.docx",
	"/j/foo/bar/a/b/c/d/foo.png",
	"/j/foo/bar/a/b/c/d/e/f/g/foo.PDF",
	"/j/note.md",
	"/j/foo/bar/a/b/c/d/e/f/g/r.txt",
	"/j/foo/bar/y/t.docx",
	"/j/foo/bar/a/b/c/d/e/f/k/r.txt",
	"/j/foo/x.go",
}

func TestGetPathsForMkdirs(t *testing.T) {
	result := getPathsForMkdirs(pathForInputs)

	var expected = []string{
		"/j/foo/bar/a/b/c/d/e/f/g",
		"/j/foo/bar/a/b/c/d/e/f/k",
		"/j/foo/bar/y",
		"/j/c",
	}

	assert.ElementsMatch(t, expected, result)
}

func TestGetPathsForRMDir(t *testing.T) {
	result := getPathsForRMDir(pathForInputs)

	var expected = []string{
		"/j/foo/bar/a",
		"/j/foo/bar/y",
		"/j/c",
	}

	assert.ElementsMatch(t, expected, result)
}
