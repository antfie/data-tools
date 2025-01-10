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
	"os"
)

type HashSignature struct {
	HashID     *uint
	Hash       string
	Size       uint
	FileTypeID *uint
	FileType   string
	fileIDs    []uint // This is not exported to prevent GORM from trying to map it
}

type FileIdAndPath struct {
	FileID       uint
	AbsolutePath string
}

func (ctx *Context) HashFiles() error {
	var count int64 = 0
	result := ctx.DB.Model(&models.File{}).Where("deleted_at IS NULL AND file_hash_id IS NULL AND size IS NULL AND file_type_id IS NULL AND ignored = 0").Count(&count)

	if result.Error != nil {
		return result.Error
	}

	// Nothing to do
	if count == 0 {
		utils.ConsoleAndLogPrintf("No files to hash. Have you already crawled?")
		return nil
	}

	utils.ConsoleAndLogPrintf("Acquiring data")

	var hashSignatures []HashSignature
	result = ctx.DB.Raw(QueryGetExistingHashSignatures()).Scan(&hashSignatures)

	if result.Error != nil {
		return result.Error
	}

	var existingFileTypes []models.FileType
	result = ctx.DB.Raw(QueryGetExistingFileTypes()).Scan(&existingFileTypes)

	if result.Error != nil {
		return result.Error
	}

	utils.ConsoleAndLogPrintf("Hashing %s", utils.Pluralize("file", count))

	bar := progressbar.Default(count)

	totalNewUniqueHashes := int64(0)
	duplicateFileHashes := 0

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
				utils.ConsoleAndLogPrintf("Total new and unique file hashes found: %s, duplicate file hashes: %s", humanize.Comma(totalNewUniqueHashes), humanize.Comma(int64(duplicateFileHashes)))
			}

			return nil
		}

		var notFoundFileIDs []uint

		orchestrator := utils.NewTaskOrchestrator(bar, len(files), ctx.Config.MaxConcurrentFileOperations)

		for _, file := range files {
			orchestrator.StartTask()
			go hashFile(orchestrator, &hashSignatures, &notFoundFileIDs, file)
		}

		orchestrator.WaitForTasks()

		err := ctx.DB.Transaction(func(tx *gorm.DB) error {
			for hashSignatureIndex, hashSignature := range hashSignatures {
				// Try to resolve the file type ID if required
				if hashSignature.FileTypeID == nil {
					for _, fileType := range existingFileTypes {
						if hashSignature.FileType == fileType.Type {
							hashSignatures[hashSignatureIndex].FileTypeID = &fileType.ID
							break
						}
					}
				}

				// Create a new FileType if required
				if hashSignatures[hashSignatureIndex].FileTypeID == nil {
					fileTypeModel := models.FileType{Type: hashSignature.FileType}

					result = tx.Create(&fileTypeModel)

					if result.Error != nil {
						return result.Error
					}

					hashSignatures[hashSignatureIndex].FileTypeID = &fileTypeModel.ID
					existingFileTypes = append(existingFileTypes, fileTypeModel)
				}

				// Create a new FileHash if required
				if hashSignature.HashID == nil {
					model := models.FileHash{
						Hash:       hashSignature.Hash,
						FileTypeID: hashSignatures[hashSignatureIndex].FileTypeID,
						Size:       &hashSignature.Size,
					}

					result = tx.Create(&model)
					totalNewUniqueHashes++

					if len(hashSignature.fileIDs) > 1 {
						duplicateFileHashes += len(hashSignature.fileIDs) - 1
					}

					if result.Error != nil {
						return result.Error
					}

					hashSignatures[hashSignatureIndex].HashID = &model.ID
				} else {
					duplicateFileHashes += len(hashSignature.fileIDs)
				}

				for _, fileID := range hashSignature.fileIDs {
					result = tx.Model(&models.File{}).Where("id = ?", fileID).Updates(models.File{
						FileHashID: hashSignatures[hashSignatureIndex].HashID,
						Size:       &hashSignature.Size,
						FileTypeID: hashSignatures[hashSignatureIndex].FileTypeID,
					})

					if result.Error != nil {
						return result.Error
					}
				}
			}

			if len(notFoundFileIDs) > 0 {
				result = tx.Where("id IN ?", notFoundFileIDs).Delete(&models.File{})

				if result.Error != nil {
					return result.Error
				}
			}

			return nil
		})

		if err != nil {
			return err
		}
	}
}

func hashFile(orchestrator *utils.TaskOrchestrator, existingHashSignatures *[]HashSignature, notFoundFileIDs *[]uint, file FileIdAndPath) {
	fileInfo, err := os.Stat(file.AbsolutePath)

	// If the file does not exist we can ignore it
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("Ignoring not-found file \"%s\"", file.AbsolutePath)

			orchestrator.Lock()
			*notFoundFileIDs = append(*notFoundFileIDs, file.FileID)
			orchestrator.Unlock()

			orchestrator.FinishTask()
			return
		}

		log.Printf("Error: Could not open file \"%s\": %v", file.AbsolutePath, err)
		orchestrator.FinishTask()
		return
	}

	// Do file typing first to fail faster if there is a file issue
	fileType, err := GetFileType(file.AbsolutePath)

	if err != nil {
		log.Printf("Error: Could not type file \"%s\": %v", file.AbsolutePath, err)

		orchestrator.FinishTask()
		return
	}

	hash, err := crypto.HashFile(file.AbsolutePath)

	if err != nil {
		log.Printf("Error: Could not hash file \"%s\": %v", file.AbsolutePath, err)

		orchestrator.FinishTask()
		return
	}

	size := fileInfo.Size()

	// Ensure we have not wrapped around for uint conversion
	if size < 0 {
		log.Printf("Error: Negative file size \"%s\"", file.AbsolutePath)

		orchestrator.FinishTask()
		return
	}

	signature := HashSignature{
		Hash:     hash,
		Size:     uint(size),
		FileType: fileType,
		fileIDs:  []uint{file.FileID},
	}

	// Maps are not threadsafe
	orchestrator.Lock()

	for existingHashSignatureIndex, existingHashSignature := range *existingHashSignatures {
		if existingHashSignature.Hash == signature.Hash {
			// Do hash collision detection on the found hash
			if existingHashSignature.Size != signature.Size {
				log.Printf("File \"%s\" has unexpected size. Expected %d, got %d. Has a hash collision occured?", file.AbsolutePath, existingHashSignature.Size, signature.Size)
				orchestrator.Unlock()

				orchestrator.FinishTask()
				return
			}

			if existingHashSignature.FileType != signature.FileType {
				log.Printf("File \"%s\" has unexpected type. Expected \"%s\", got \"%s\". Has a hash collision occured?", file.AbsolutePath, existingHashSignature.FileType, signature.FileType)
				orchestrator.Unlock()

				orchestrator.FinishTask()
				return
			}

			(*existingHashSignatures)[existingHashSignatureIndex].fileIDs = append(existingHashSignature.fileIDs, file.FileID)
			orchestrator.Unlock()

			orchestrator.FinishTask()
			return
		}
	}

	*existingHashSignatures = append(*existingHashSignatures, signature)
	orchestrator.Unlock()

	orchestrator.FinishTask()
}
