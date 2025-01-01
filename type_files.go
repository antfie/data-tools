package main

import (
	"data-tools-2025/models"
	"data-tools-2025/utils"
	"errors"
	"github.com/schollz/progressbar/v3"
	"gorm.io/gorm"
	"log"
)

func (ctx *Context) TypeFiles() error {
	var count int64 = 0
	result := ctx.DB.Model(&models.FileHash{}).Where("file_type_id IS NULL AND ignored = 0").Count(&count)

	if result.Error != nil {
		return result.Error
	}

	// Nothing to do
	if count == 0 {
		log.Print("No files to type. Have you already hashed?")
		return nil
	}

	utils.ConsoleAndLogPrintf("Typing %s with %s in batches of %d", utils.Pluralize("file", count), utils.Pluralize("thread", ctx.Config.MaxConcurrentFileOperations), ctx.Config.BatchSize)

	bar := progressbar.Default(count)

	// Do batches until there are no more
	for {
		var fileHashesToType []FileIdAndPath
		result = ctx.DB.Raw(QueryUnTypedFileHashesWithLimit(), ctx.Config.BatchSize).Scan(&fileHashesToType)

		if result.Error != nil {
			return result.Error
		}

		// Have we finished?
		if fileHashesToType == nil {

			// Update the file types from the file hash types
			result = ctx.DB.Exec(`UPDATE files
			SET file_type_id = fh.file_type_id
			FROM file_hashes fh
			WHERE files.file_hash_id = fh.id`)

			return result.Error
		}

		fileHashTypeMap := make(map[string][]uint, len(fileHashesToType))
		var uniqueFileTypes []string

		orchestrator := utils.NewTaskOrchestrator(bar, len(fileHashesToType), ctx.Config.MaxConcurrentFileOperations)

		for _, fileHash := range fileHashesToType {
			orchestrator.StartTask()
			go typeFile(orchestrator, fileHash, fileHashTypeMap, &uniqueFileTypes)
		}

		orchestrator.WaitForTasks()

		// Find existing types in the db
		var existingDBFileTypes []models.FileType
		result = ctx.DB.Where("type IN ?", uniqueFileTypes).Find(&existingDBFileTypes)

		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}

		var newDBFileTypes []models.FileType
		for _, fileType := range uniqueFileTypes {
			if getDBFileType(fileType, existingDBFileTypes) == nil {
				newDBFileTypes = append(newDBFileTypes, models.FileType{Type: fileType})
			}
		}

		if len(newDBFileTypes) > 0 {
			// Insert new file types if required
			result = ctx.DB.Create(&newDBFileTypes)

			if result.Error != nil {
				return result.Error
			}

			existingDBFileTypes = append(existingDBFileTypes, newDBFileTypes...)
		}

		for fileType, fileIds := range fileHashTypeMap {
			dbFileType := getDBFileType(fileType, existingDBFileTypes)

			if dbFileType == nil {
				return ErrCouldNotResolveFileType
			}

			result = ctx.DB.Model(&models.FileHash{}).Where("id IN ?", fileIds).Updates(models.FileHash{
				FileTypeID: &dbFileType.ID,
			})

			if result.Error != nil {
				return result.Error
			}
		}
	}
}

func typeFile(orchestrator *utils.TaskOrchestrator, fileHashToType FileIdAndPath, fileHashTypeMap map[string][]uint, uniqueFileTypes *[]string) {
	fileType, err := GetTypeOfFile(fileHashToType.AbsolutePath)

	if err != nil {
		log.Fatalf("Could not type file \"%s\": %v", fileHashToType.AbsolutePath, err)
	}

	// Maps are not threadsafe
	orchestrator.Lock()
	existingFileHashIdsWithThisType, found := fileHashTypeMap[fileType]

	if !found {
		fileHashTypeMap[fileType] = []uint{fileHashToType.ID}
		*uniqueFileTypes = append(*uniqueFileTypes, fileType)
	} else {
		fileHashTypeMap[fileType] = append(existingFileHashIdsWithThisType, fileHashToType.ID)
	}
	orchestrator.Unlock()

	orchestrator.FinishTask()
}

func getDBFileType(fileType string, fileTypes []models.FileType) *models.FileType {
	for _, dbFileType := range fileTypes {
		if fileType == dbFileType.Type {
			return &dbFileType
		}
	}

	return nil
}
