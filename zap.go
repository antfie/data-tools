package main

import (
	"data-tools/models"
	"data-tools/utils"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/dustin/go-humanize"
	"github.com/schollz/progressbar/v3"
	"log"
	"os"
	"path"
	"path/filepath"
)

type ZapResult struct {
	FileHashID   uint
	Hash         string
	FileID       uint
	AbsolutePath string
}

func (ctx *Context) Zap(outputPath string, safeMode bool) error {
	type ZapInfo struct {
		FileHashesToZap         int64
		UniqueHashTotalFileSize uint64
		TotalFileSize           uint64
	}

	var info ZapInfo
	result := ctx.DB.Raw(`
SELECT (SELECT COUNT(*) FROM file_hashes WHERE size IS NOT NULL AND ignored = 0 AND zapped = 0) file_hashes_to_zap,
       (SELECT SUM(size) FROM file_hashes WHERE size IS NOT NULL AND ignored = 0 AND zapped = 0) unique_hash_total_file_size,
       (SELECT SUM(size) FROM files WHERE deleted_at IS NULL AND size IS NOT NULL AND ignored = 0 AND zapped = 0) total_file_size
`).First(&info)

	if result.Error != nil {
		return result.Error
	}

	// Nothing to do
	if info.FileHashesToZap == 0 {
		utils.ConsoleAndLogPrintf("No files to ZAP. Have you already sized?")
		return nil
	}

	outputPathAbs, err := filepath.Abs(outputPath)

	if err != nil {
		return err
	}

	remainingPercentage := (float64(info.TotalFileSize-info.UniqueHashTotalFileSize) / float64(info.TotalFileSize)) * 100
	removalPercentage := (float64(info.UniqueHashTotalFileSize) / float64(info.TotalFileSize)) * 100
	utils.ConsoleAndLogPrintf("Zapping (de-duplicating) %s (%s) to \"%s\". This is %.2f%% of %s, a reduction of %s (%.2f%%)", utils.Pluralize("file", info.FileHashesToZap), humanize.Bytes(info.TotalFileSize-info.UniqueHashTotalFileSize), outputPathAbs, remainingPercentage, humanize.Bytes(info.TotalFileSize), humanize.Bytes(info.TotalFileSize-(info.TotalFileSize-info.UniqueHashTotalFileSize)), removalPercentage)

	err = os.MkdirAll(outputPathAbs, 0700)

	if err != nil {
		return err
	}

	bar := progressbar.Default(info.FileHashesToZap)

	// Do batches until there are no more
	for {
		var fileHashesToZap []ZapResult
		result = ctx.DB.Raw(QueryGetFileHashesToZapWithLimit(), ctx.Config.BatchSize).Scan(&fileHashesToZap)

		if result.Error != nil {
			return result.Error
		}

		// Have we finished?
		if fileHashesToZap == nil {
			// Update the file zap status from the file hash zap status
			result = ctx.DB.Exec(`UPDATE files
			SET zapped = 1
			FROM file_hashes fh
			WHERE files.file_hash_id = fh.id
			AND	fh.zapped = 1
		  	AND fh.ignored = 0
			AND	files.zapped = 0
			AND files.deleted_at IS NULL
			AND files.ignored = 0`)

			return result.Error
		}

		var notFoundFileIDs []uint

		orchestrator := utils.NewTaskOrchestrator(bar, len(fileHashesToZap), ctx.Config.MaxConcurrentFileOperations)

		for _, fileHash := range fileHashesToZap {
			orchestrator.StartTask()
			go ctx.zapFile(orchestrator, safeMode, outputPathAbs, fileHash, &notFoundFileIDs)
		}

		orchestrator.WaitForTasks()

		if !safeMode {
			// TODO: Bulk remove duplicates or bin off the root folder. Maybe better doing this in zapFile for safery
		}
	}
}

func (ctx *Context) zapFile(orchestrator *utils.TaskOrchestrator, safeMode bool, zapBasePath string, file ZapResult, notFoundFileIDs *[]uint) {
	// If the file does not exist we can ignore it
	if !IsFile(file.AbsolutePath) {
		orchestrator.Lock()
		log.Printf("Ignoring not-found file \"%s\"", file.AbsolutePath)
		*notFoundFileIDs = append(*notFoundFileIDs, file.FileID)
		orchestrator.Unlock()

		orchestrator.FinishTask()
		return
	}

	// Store as hex so it will work OK on case-insensitive filesystems
	hexFileName := hex.EncodeToString(base58.Decode(file.Hash))

	// Only move if not in safe mode
	move := !safeMode

	// ZAP
	err := CopyOrMoveFile(file.AbsolutePath, path.Join(zapBasePath, hexFileName), move)

	if err != nil {
		log.Fatalf("Could not ZAP file \"%s\": %v", file.AbsolutePath, err)
	}

	result := ctx.DB.Model(&models.FileHash{}).Where("id = ?", file.FileHashID).Updates(models.FileHash{
		Zapped: true,
	})

	if result.Error != nil {
		log.Fatalf("DB Error: %v", result.Error)
	}

	if !safeMode {
		// TODO: Remove any other duplicates, or do that in bulk?
		// might be safer to do this here
	}

	orchestrator.FinishTask()
}
