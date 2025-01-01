package crypto

import "testing"
import "github.com/stretchr/testify/assert"

func TestHashFile(t *testing.T) {
	result, err := HashFile("../test/data/a/file.md")
	assert.NoError(t, err)

	expected := "3tSamSfZTrePjU1wwBcwGjo1tGujGVoAjcPAt6mis6Adr5jMUFQZPY2dBVRV4RKX5UReejgzZdkTEQVFTqjBVjVq"
	assert.Equal(t, expected, result)
}
