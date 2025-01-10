package main

import (
	"data-tools/models"
	"data-tools/utils"
	"errors"
	"github.com/dustin/go-humanize"
	"github.com/schollz/progressbar/v3"
	"log"
	"os"
)

type FileHashAndFile struct {
	FileHashID   uint
	FileID       uint
	AbsolutePath string
}

type FileIdAndPath struct {
	FileID       uint
	AbsolutePath string
}

func (ctx *Context) SizeFiles() error {
	var count int64 = 0
	result := ctx.DB.Model(&models.FileHash{}).Where("size IS NULL AND ignored = 0").Count(&count)

	if result.Error != nil {
		return result.Error
	}

	// Nothing to do
	if count == 0 {
		utils.ConsoleAndLogPrintf("No files to size. Have you already hashed?")
		return nil
	}

	utils.ConsoleAndLogPrintf("Sizing %s", utils.Pluralize("file", count))

	bar := progressbar.Default(count)

	totalFilesSized := int64(0)
	totalFileSize := uint(0)

	// Do batches until there are no more
	for {
		var fileHashesToSize []FileHashAndFile
		result = ctx.DB.Raw(QueryUnSizedFileHashesWithLimit(), ctx.Config.BatchSize).Scan(&fileHashesToSize)

		if result.Error != nil {
			return result.Error
		}

		// Have we finished?
		if fileHashesToSize == nil {
			if totalFilesSized > 0 {
				utils.ConsoleAndLogPrintf("Sized %s totalling: %s", utils.Pluralize("file", totalFilesSized), humanize.Bytes(uint64(totalFileSize)))
			}

			return nil
		}

		var notFoundFileIDs []uint

		orchestrator := utils.NewTaskOrchestrator(bar, len(fileHashesToSize), ctx.Config.MaxConcurrentFileOperations)

		for _, fileHash := range fileHashesToSize {
			orchestrator.StartTask()
			go ctx.sizeFile(orchestrator, fileHash, &notFoundFileIDs, &totalFilesSized, &totalFileSize)
		}

		orchestrator.WaitForTasks()

		// Update the file sizes from the file hash sizes
		result = ctx.DB.Exec(`UPDATE files
			SET size = fh.size
			FROM file_hashes fh
			WHERE files.file_hash_id = fh.id
			AND fh.size IS NOT NULL
			AND fh.ignored = 0
			AND files.size IS NULL
			AND files.deleted_at IS NULL
			AND files.ignored = 0`)

		if result.Error != nil {
			return result.Error
		}

		if len(notFoundFileIDs) > 0 {
			result = ctx.DB.Where("id IN ?", notFoundFileIDs).Delete(&models.File{})

			if result.Error != nil {
				return result.Error
			}
		}
	}
}

func (ctx *Context) sizeFile(orchestrator *utils.TaskOrchestrator, file FileHashAndFile, notFoundFileIDs *[]uint, totalFilesSized *int64, totalFileSize *uint) {
	info, err := os.Stat(file.AbsolutePath)

	if err != nil {
		// If the file does not exist we can ignore it
		if errors.Is(err, os.ErrNotExist) {
			orchestrator.Lock()
			log.Printf("Ignoring not-found file \"%s\"", file.AbsolutePath)
			*notFoundFileIDs = append(*notFoundFileIDs, file.FileID)
			orchestrator.Unlock()

			orchestrator.FinishTask()
			return
		}
		
		log.Printf("Could not open file \"%s\": %v", file.AbsolutePath, err)
		orchestrator.FinishTask()
		return
	}

	size := info.Size()

	// Ensure we have not wrapped around for uint conversion
	if size < 0 {
		log.Fatalf("Negative file size")
	}

	formattedSize := uint(size)

	result := ctx.DB.Model(&models.FileHash{}).Where("id = ?", file.FileHashID).Updates(models.FileHash{
		Size: &formattedSize,
	})

	if result.Error != nil {
		log.Fatalf("DB Error: %v", result.Error)
	}

	orchestrator.Lock()
	*totalFilesSized++
	*totalFileSize += formattedSize
	orchestrator.Unlock()

	orchestrator.FinishTask()
}
