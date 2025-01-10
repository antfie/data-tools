package main

import (
	"data-tools/models"
	"data-tools/utils"
	"errors"
	"github.com/schollz/progressbar/v3"
	"log"
	"os"
	"strings"
)

type Sanity struct {
	FileID       uint
	Hash         string
	Size         int64
	Type         string
	AbsolutePath string
}

func (ctx *Context) DuplicateHashSanityCheck() error {
	utils.ConsoleAndLogPrintf("Acquiring data")

	var hashesToProcess []string
	result := ctx.DB.Raw(QueryHashSanity()).Scan(&hashesToProcess)

	if result.Error != nil {
		return result.Error
	}

	fileCount, batchesOfFileIdsToProcess := createBatches(hashesToProcess, ctx)

	// Nothing to do
	if fileCount == 0 {
		utils.ConsoleAndLogPrintf("No files to sanity check. Have you already hashed, sized and typed?")
		return nil
	}

	utils.ConsoleAndLogPrintf("Sanity checking %s", utils.Pluralize("file", fileCount))

	bar := progressbar.Default(fileCount)

	for _, batch := range batchesOfFileIdsToProcess {
		var filesToCheck []Sanity
		result := ctx.DB.Raw(QueryFileForHashSanityByIDs(), batch).Scan(&filesToCheck)

		if result.Error != nil {
			return result.Error
		}

		var notFoundFileIDs []uint

		orchestrator := utils.NewTaskOrchestrator(bar, len(filesToCheck), ctx.Config.MaxConcurrentFileOperations)

		for _, file := range filesToCheck {
			orchestrator.StartTask()
			go ctx.sanityCheckFile(orchestrator, file, &notFoundFileIDs)
		}

		orchestrator.WaitForTasks()

		if len(notFoundFileIDs) > 0 {
			result = ctx.DB.Where("id IN ?", notFoundFileIDs).Delete(&models.File{})

			if result.Error != nil {
				return result.Error
			}
		}
	}

	return nil
}

func createBatches(hashesToProcess []string, ctx *Context) (int64, [][]string) {
	count := int64(0)
	var batchesOfFileIdsToProcess [][]string
	batchIndex := -1
	batchCount := int64(-1)

	for _, hash := range hashesToProcess {
		for _, fileId := range strings.Split(hash, ",") {
			if batchCount == -1 || batchCount > ctx.Config.BatchSize-1 {
				batchCount = 0
				batchIndex++
				batchesOfFileIdsToProcess = append(batchesOfFileIdsToProcess, []string{fileId})

			} else {
				batchesOfFileIdsToProcess[batchIndex] = append(batchesOfFileIdsToProcess[batchIndex], fileId)
			}

			count++
			batchCount++
		}
	}

	return count, batchesOfFileIdsToProcess
}

func (ctx *Context) sanityCheckFile(orchestrator *utils.TaskOrchestrator, file Sanity, notFoundFileIDs *[]uint) {
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

		log.Printf("Error: Could not open file \"%s\": %v", file.AbsolutePath, err)
		orchestrator.FinishTask()
		return
	}

	size := info.Size()

	if size != file.Size {
		log.Printf("File \"%s\" has incorrect size. Expected %d, got %d", file.AbsolutePath, file.Size, size)
	}

	fileType, err := GetFileType(file.AbsolutePath)

	if err != nil {
		log.Fatalf("Could not type file \"%s\": %v", file.AbsolutePath, err)
	}

	if fileType != file.Type {
		log.Printf("File \"%s\" has incorrect type. Expected %s, got %s", file.AbsolutePath, file.Type, fileType)
	}

	orchestrator.FinishTask()
}
