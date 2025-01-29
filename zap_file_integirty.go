package main

import (
	"data-tools/models"
	"errors"
	"log"
	"os"
	"path"
)

// TODO: file integrity check on the zap folder. Do the hashes match the filenames

// first check sizes with os.Stat call against DB, first pass is this.
// Then we can check the actual hashes

func (ctx *Context) ZapDBIntegrityTestBySize() error {
	hashes := make(map[string]int64)

	// Execute the raw SQL query
	rows, err := ctx.DB.Raw(`
SELECT		hash,
			size
FROM 		file_hashes
WHERE		zapped = 1
AND			size IS NOT NULL
AND			ignored = 0
ORDER BY    id -- for deterministic result order
`).Rows()
	if err != nil {
		return err
	}

	for rows.Next() {
		var hash string
		var size int64
		if rowErr := rows.Scan(&hash, &size); rowErr != nil {
			return err
		}
		hashes[hash] = size
	}

	err = rows.Close()

	if err != nil {
		return err
	}

	notFoundHashes, err := AssertHashesInZapPath(ctx.Config.ZapDataPath, hashes)

	if err != nil {
		return err
	}

	if len(notFoundHashes) > 0 {
		// Urgh!
		// NOTE When update with struct, GORM will only update non-zero fields, you might want to use map to update attributes or use Select to specify fields to update
		// - https://gorm.io/docs/update.html#Updates-multiple-columns
		//result := ctx.DB.Model(models.FileHash{}).Where("hash IN ?", notFoundHashes).Updates(map[string]interface{}{"zapped": false})
		result := ctx.DB.Model(models.FileHash{}).Where("hash IN ?", notFoundHashes).Update("zapped", false)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected != int64(len(notFoundHashes)) {
			return errors.New("unable to update hashes")
		}
	}

	return nil
}

func AssertHashesInZapPath(zapPath string, hashes map[string]int64) ([]string, error) {
	var notFoundHashes []string

	for hash, size := range hashes {
		filePath := path.Join(zapPath, FormatRelativeZapFilePathFromEncodedHash(hash))

		stat, err := os.Stat(filePath)

		if os.IsNotExist(err) {
			log.Printf("Hash not found in ZAP folder: %s", hash)
			notFoundHashes = append(notFoundHashes, hash)
		} else if err != nil {
			return nil, err
		}

		if stat.Size() != size {
			log.Printf("Hash size mismatch: expected %d, got %d for hash %s", size, stat.Size(), hash)
			notFoundHashes = append(notFoundHashes, hash)
		}
	}

	return notFoundHashes, nil
}
