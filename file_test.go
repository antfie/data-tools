package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPathsForMkdirs(t *testing.T) {
	var inputs = []string{
		"/c/t.docx",
		"/foo/bar/a/b/c/d/foo.png",
		"/foo/bar/a/b/c/d/e/f/g/foo.PDF",
		"/note.md",
		"/foo/bar/a/b/c/d/e/f/g/r.txt",
		"/foo/bar/y/t.docx",
		"/foo/bar/a/b/c/d/e/f/k/r.txt",
		"/foo/x.go",
	}

	result := getPathsForMkdirs(inputs)

	var expected = []string{
		"/foo/bar/a/b/c/d/e/f/g",
		"/foo/bar/a/b/c/d/e/f/k",
		"/foo/bar/y",
		"/c",
	}

	assert.ElementsMatch(t, expected, result)
}
