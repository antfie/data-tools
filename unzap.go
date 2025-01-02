package main

import (
	"data-tools/models"
	"data-tools/utils"
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/dustin/go-humanize"
	"github.com/schollz/progressbar/v3"
	"log"
	"os"
	"path"
	"path/filepath"
)

func (ctx *Context) UnZap(sourcePath, outputPath string) error {
	_, err := os.Stat(outputPath)

	// We expect the output directory to be empty
	if err != nil || !errors.Is(err, os.ErrNotExist) {
		return ErrDestinationPathNotEmpty
	}

	type ZapInfo struct {
		ZappedFileHashes        int64
		UniqueHashTotalFileSize uint64
		TotalFileSize           uint64
	}

	var info ZapInfo
	result := ctx.DB.Raw(`
SELECT (SELECT COUNT(*) FROM file_hashes WHERE size IS NOT NULL AND ignored = 0 AND zapped = 1) zapped_file_hashes,
       (SELECT SUM(size) FROM file_hashes WHERE size IS NOT NULL AND ignored = 0 AND zapped = 1) unique_hash_total_file_size,
       (SELECT SUM(size) FROM files WHERE deleted_at IS NULL AND size IS NOT NULL AND ignored = 0 AND zapped = 1) total_file_size
`).First(&info)

	if result.Error != nil {
		return result.Error
	}

	// Nothing to do
	if info.ZappedFileHashes == 0 {
		log.Print("No files to un-ZAP. Have you already ZAPped?")
		return nil
	}

	destinationAbsolutePath, err := filepath.Abs(outputPath)

	if err != nil {
		return err
	}

	percentage := 100 - ((float64(info.TotalFileSize-info.UniqueHashTotalFileSize) / float64(info.TotalFileSize)) * 100)
	utils.ConsoleAndLogPrintf("Un-ZAPing %s to %s (%.2f%%) at \"%s\"", humanize.Bytes(info.TotalFileSize-info.UniqueHashTotalFileSize), humanize.Bytes(info.TotalFileSize), percentage, destinationAbsolutePath)

	bar := progressbar.Default(info.ZappedFileHashes)

	// We need something in the array for gorm to work. No file will have an ID of 0
	processedFileIds := []uint{0}

	// Do batches until there are no more
	for {
		var fileHashesToUnZap []ZapResult
		result = ctx.DB.Raw(QueryGetZappedFileHashesToUnZapWithLimit(), processedFileIds, ctx.Config.BatchSize).Scan(&fileHashesToUnZap)

		if result.Error != nil {
			return result.Error
		}

		// Have we finished?
		if fileHashesToUnZap == nil {
			return nil
		}

		var notFoundFileIDs []uint

		orchestrator := utils.NewTaskOrchestrator(bar, len(fileHashesToUnZap), ctx.Config.MaxConcurrentFileOperations)

		zapSourcePath := path.Join(sourcePath, zapFolderName)

		for _, fileHash := range fileHashesToUnZap {
			orchestrator.StartTask()
			go ctx.unZapFile(orchestrator, &processedFileIds, zapSourcePath, destinationAbsolutePath, &fileHash, &notFoundFileIDs)
		}

		orchestrator.WaitForTasks()

		if len(notFoundFileIDs) > 0 {
			result = ctx.DB.Where("id IN ?", notFoundFileIDs).Delete(&models.File{})

			if result.Error != nil {
				return result.Error
			}
		}
	}
}

func (ctx *Context) unZapFile(orchestrator *utils.TaskOrchestrator, processedFileIds *[]uint, zapSourcePath, destinationAbsolutePath string, file *ZapResult, notFoundFileIDs *[]uint) {
	// If the file does not exist we can ignore it
	if !IsFile(zapSourcePath) {
		orchestrator.Lock()
		log.Printf("Ignoring not-found file \"%s\"", file.AbsolutePath)
		*notFoundFileIDs = append(*notFoundFileIDs, file.FileID)
		orchestrator.Unlock()

		orchestrator.FinishTask()
		return
	}

	destinationPath := path.Join(destinationAbsolutePath, file.AbsolutePath)

	// Make the destination directory if required
	err := os.MkdirAll(path.Dir(destinationPath), 0700)

	if err != nil {
		log.Panic(err)
	}

	hexFileName := hex.EncodeToString(base58.Decode(file.Hash))

	// un-ZAP
	err = CopyOrMoveFile(path.Join(zapSourcePath, hexFileName), destinationPath, false)

	if err != nil {
		log.Panic(err)
	}

	orchestrator.Lock()
	*processedFileIds = append(*processedFileIds, file.FileID)
	orchestrator.Unlock()

	orchestrator.FinishTask()
}