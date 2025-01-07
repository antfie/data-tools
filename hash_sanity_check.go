package main

import (
	"data-tools/models"
	"data-tools/utils"
	"errors"
	"github.com/schollz/progressbar/v3"
	"log"
	"os"
)

type Sanity struct {
	FileID       uint
	FileHashID   uint
	Hash         string
	Size         int64
	Type         string
	AbsolutePath string
}

func (ctx *Context) DuplicateHashSanityCheck() error {
	var filesToCheck []Sanity
	result := ctx.DB.Raw(QueryHashSanity(), ctx.Config.BatchSize).Scan(&filesToCheck)

	if result.Error != nil {
		return result.Error
	}

	// Nothing to do
	if len(filesToCheck) == 0 {
		utils.ConsoleAndLogPrintf("No files to check. Have you already hashed, sized and typed?")
		return nil
	}

	count := int64(len(filesToCheck))

	utils.ConsoleAndLogPrintf("Sanity checking %s hashes", utils.Pluralize("file", count))

	bar := progressbar.Default(count)

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

	return nil
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
		} else {
			log.Fatalf("Could not open file \"%s\": %v", file.AbsolutePath, err)
		}
	}

	size := info.Size()

	if size != file.Size {
		log.Printf("File \"%s\" has incorrect size. Expected %d, got %d", file.AbsolutePath, file.Size, size)
	}

	fileType, err := GetTypeOfFile(file.AbsolutePath)

	if err != nil {
		log.Fatalf("Could not type file \"%s\": %v", file.AbsolutePath, err)
	}

	if fileType != file.Type {
		log.Printf("File \"%s\" has incorrect type. Expected %s, got %s", file.AbsolutePath, file.Type, fileType)
	}

	orchestrator.FinishTask()
}
