package main

import (
	"data-tools/crypto"
	"data-tools/models"
	"data-tools/utils"
	"errors"
	"github.com/dustin/go-humanize"
	"github.com/schollz/progressbar/v3"
	"gorm.io/gorm"
	"log"
)

func (ctx *Context) HashFiles() error {
	var count int64 = 0
	result := ctx.DB.Model(&models.File{}).Where("deleted_at IS NULL AND file_hash_id IS NULL AND ignored = 0").Count(&count)

	if result.Error != nil {
		return result.Error
	}

	// Nothing to do
	if count == 0 {
		utils.ConsoleAndLogPrintf("No files to hash. Have you already crawled?")
		return nil
	}

	utils.ConsoleAndLogPrintf("Hashing %s", utils.Pluralize("file", count))

	bar := progressbar.Default(count)

	totalNewUniqueHashes := int64(0)
	duplicateHashes := int64(0)

	// Do batches until there are no more
	for {
		var files []FileIdAndPath
		result := ctx.DB.Raw(QueryUnHashedFilePathsWithLimit(), ctx.Config.BatchSize).Scan(&files)

		if result.Error != nil {
			return result.Error
		}

		// Have we finished?
		if files == nil {
			if totalNewUniqueHashes > 0 {
				utils.ConsoleAndLogPrintf("Total new and unique hashes found: %d, duplicate hashes found: %d", humanize.Comma(totalNewUniqueHashes), humanize.Comma(duplicateHashes))
			}

			return nil
		}

		hashes := make(map[string][]uint, len(files))
		var uniqueHashes []string
		var notFoundFileIDs []uint

		orchestrator := utils.NewTaskOrchestrator(bar, len(files), ctx.Config.MaxConcurrentFileOperations)

		for _, file := range files {
			orchestrator.StartTask()
			go hashFile(orchestrator, hashes, &uniqueHashes, &notFoundFileIDs, file)
		}

		orchestrator.WaitForTasks()

		// Find existing hashes in the db
		var existingDBHashes []models.FileHash
		result = ctx.DB.Where("hash IN ?", uniqueHashes).Find(&existingDBHashes)

		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}

		var newDBHashes []models.FileHash
		for _, hash := range uniqueHashes {
			if getDBHash(hash, existingDBHashes) == nil {
				totalNewUniqueHashes++
				newDBHashes = append(newDBHashes, models.FileHash{Hash: hash})
			} else {
				duplicateHashes++
			}
		}

		if len(newDBHashes) > 0 {
			// Insert new hashes if required
			result = ctx.DB.Create(&newDBHashes)

			if result.Error != nil {
				return result.Error
			}

			existingDBHashes = append(existingDBHashes, newDBHashes...)
		}

		for hash, fileIds := range hashes {
			dbHash := getDBHash(hash, existingDBHashes)

			if dbHash == nil {
				return ErrCouldNotResolveHash
			}

			result = ctx.DB.Model(&models.File{}).Where("id IN ?", fileIds).Updates(models.File{
				FileHashID: &dbHash.ID,
			})

			if result.Error != nil {
				return result.Error
			}
		}

		if len(notFoundFileIDs) > 0 {
			result = ctx.DB.Where("id IN ?", notFoundFileIDs).Delete(&models.File{})

			if result.Error != nil {
				return result.Error
			}
		}
	}
}

func hashFile(orchestrator *utils.TaskOrchestrator, hashes map[string][]uint, uniqueHashes *[]string, notFoundFileIDs *[]uint, file FileIdAndPath) {
	// If the file does not exist we can ignore it
	if !IsFile(file.AbsolutePath) {
		orchestrator.Lock()
		log.Printf("Ignoring not-found file \"%s\"", file.AbsolutePath)
		*notFoundFileIDs = append(*notFoundFileIDs, file.FileID)
		orchestrator.Unlock()

		orchestrator.FinishTask()
		return
	}

	hash, err := crypto.HashFile(file.AbsolutePath)

	if err != nil {
		log.Fatalf("Could not hash file \"%s\": %v", file.AbsolutePath, err)
	}

	// Maps are not threadsafe
	orchestrator.Lock()
	existingFileIdsWithThisHash, found := hashes[hash]

	if !found {
		hashes[hash] = []uint{file.FileID}
		*uniqueHashes = append(*uniqueHashes, hash)
	} else {
		hashes[hash] = append(existingFileIdsWithThisHash, file.FileID)
	}
	orchestrator.Unlock()

	orchestrator.FinishTask()
}

func getDBHash(hash string, hashes []models.FileHash) *models.FileHash {
	for _, dbHash := range hashes {
		if hash == dbHash.Hash {
			return &dbHash
		}
	}

	return nil
}
