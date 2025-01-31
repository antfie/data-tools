package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPluralize(t *testing.T) {
	assert.Equal(t, "0 batches", Pluralize("batch", 0))
	assert.Equal(t, "1 batch", Pluralize("batch", 1))
	assert.Equal(t, "2 batches", Pluralize("batch", 2))
	assert.Equal(t, "0 files", Pluralize("file", 0))
	assert.Equal(t, "1 file", Pluralize("file", 1))
	assert.Equal(t, "2 files", Pluralize("file", 2))
	assert.Equal(t, "0 hashes", Pluralize("hash", 0))
	assert.Equal(t, "1 hash", Pluralize("hash", 1))
	assert.Equal(t, "2 hashes", Pluralize("hash", 2))
}
