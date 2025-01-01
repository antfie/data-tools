package main

import (
	"data-tools-2025/models"
	"data-tools-2025/utils"
	"errors"
	"github.com/schollz/progressbar/v3"
	"log"
	"os"
)

type FileIdAndPath struct {
	ID           uint
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

	utils.ConsoleAndLogPrintf("Sizing %s with %s in batches of %d", utils.Pluralize("file", count), utils.Pluralize("thread", ctx.Config.MaxConcurrentFileOperations), ctx.Config.BatchSize)

	bar := progressbar.Default(count)

	// Do batches until there are no more
	for {
		var fileHashesToSize []FileIdAndPath
		result = ctx.DB.Raw(QueryUnSizedFileHashesWithLimit(), ctx.Config.BatchSize).Scan(&fileHashesToSize)

		if result.Error != nil {
			return result.Error
		}

		// Have we finished?
		if fileHashesToSize == nil {

			// Update the file sizes from the file hash sizes
			result = ctx.DB.Exec(`UPDATE files
			SET size = fh.size
			FROM file_hashes fh
			WHERE files.file_hash_id = fh.id`)

			return result.Error
		}

		orchestrator := utils.NewTaskOrchestrator(bar, len(fileHashesToSize), ctx.Config.MaxConcurrentFileOperations)

		for _, fileHash := range fileHashesToSize {
			orchestrator.StartTask()
			go ctx.sizeFile(orchestrator, fileHash)
		}

		orchestrator.WaitForTasks()
	}
}

func (ctx *Context) sizeFile(orchestrator *utils.TaskOrchestrator, fileHashToSize FileIdAndPath) {
	info, err := os.Stat(fileHashToSize.AbsolutePath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			//orchestrator.Logf("Could not open file \"%s\": %v", fileHashToSize.AbsolutePath, err)
			log.Printf("Could not open file \"%s\": %v", fileHashToSize.AbsolutePath, err)
			orchestrator.FinishTask()
			return
		} else {
			log.Fatalf("Could not open file \"%s\": %v", fileHashToSize.AbsolutePath, err)
		}
	}

	size := info.Size()

	// Ensure we have not wrapped around for uint conversion
	if size < 0 {
		log.Fatalf("Negative file size")
	}

	formattedSize := uint(size)

	result := ctx.DB.Model(&models.FileHash{}).Where("id = ?", fileHashToSize.ID).Updates(models.FileHash{
		Size: &formattedSize,
	})

	if result.Error != nil {
		log.Fatalf("DB Error: %v", result.Error)
	}

	orchestrator.FinishTask()
}
