package main

import (
	"data-tools/config"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestGetBatches(t *testing.T) {
	tempTestDataPath := createTempTestDataPath(t)
	defer os.RemoveAll(tempTestDataPath)

	c := &config.Config{
		DBPath:                      path.Join(tempTestDataPath, "db.db"),
		BatchSize:                   2,
		MaxConcurrentFileOperations: 2,
		IsDebug:                     true,
	}

	ctx := &Context{
		Config: c,
		DB:     initDb(c),
	}

	err := ctx.Crawl(path.Join(tempTestDataPath, "a"))
	assert.NoError(t, err)

	total, batches, err := ctx.GetBatchesOfIDs(`
SELECT		id,
			BATCH_NUMBER
FROM 		files
ORDER BY	id`, "")

	assert.NoError(t, err)
	assert.NotEmpty(t, batches)
	assert.Len(t, batches, 3)
	assert.Equal(t, int64(5), total)
	assert.Equal(t, []int{1, 2}, batches[0])
	assert.Equal(t, []int{3, 4}, batches[1])
	assert.Equal(t, []int{5}, batches[2])
}
