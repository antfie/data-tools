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

const zapFolderName = "ZAP"

type ZapResult struct {
	ID           uint
	Hash         string
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

	percentage := (float64(info.TotalFileSize-info.UniqueHashTotalFileSize) / float64(info.TotalFileSize)) * 100
	utils.ConsoleAndLogPrintf("ZAPing %s (%.2f%%) of %s", humanize.Bytes(info.TotalFileSize-info.UniqueHashTotalFileSize), percentage, humanize.Bytes(info.TotalFileSize))

	zapBasePath := path.Join(outputPath, zapFolderName)
	err := os.MkdirAll(zapBasePath, 0700)

	if err != nil {
		return err
	}

	utils.ConsoleAndLogPrintf("ZAPing %s with %s in batches of %d", utils.Pluralize("file", info.FileHashesToZap), utils.Pluralize("thread", ctx.Config.MaxConcurrentFileOperations), ctx.Config.BatchSize)

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
			// Update the file sizes from the file hash sizes
			result = ctx.DB.Exec(`UPDATE files
			SET zapped = 1
			FROM file_hashes fh
			WHERE files.file_hash_id = fh.id
			AND	fh.zapped = 1`)

			return result.Error
		}

		orchestrator := utils.NewTaskOrchestrator(bar, len(fileHashesToZap), ctx.Config.MaxConcurrentFileOperations)

		for _, fileHash := range fileHashesToZap {
			orchestrator.StartTask()
			go ctx.zapHash(orchestrator, safeMode, zapBasePath, fileHash)
		}

		orchestrator.WaitForTasks()
	}
}

func (ctx *Context) zapHash(orchestrator *utils.TaskOrchestrator, safeMode bool, zapBasePath string, fileHashToZap ZapResult) {
	// Store as hex so it will work OK on case-insensitive filesystems
	hexFileName := hex.EncodeToString(base58.Decode(fileHashToZap.Hash))

	// Only move if not in safe mode
	move := !safeMode

	// ZAP
	err := CopyOrMoveFile(fileHashToZap.AbsolutePath, path.Join(zapBasePath, hexFileName), move)

	if err != nil {
		log.Fatalf("Could not ZAP file \"%s\": %v", fileHashToZap.AbsolutePath, err)
	}

	result := ctx.DB.Model(&models.FileHash{}).Where("id = ?", fileHashToZap.ID).Updates(models.FileHash{
		Zapped: true,
	})

	if result.Error != nil {
		log.Fatalf("DB Error: %v", result.Error)
	}

	orchestrator.FinishTask()
}

func (ctx *Context) UnZap(sourcePath, outputPath string) error {
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

		orchestrator := utils.NewTaskOrchestrator(bar, len(fileHashesToUnZap), ctx.Config.MaxConcurrentFileOperations)

		zapSourcePath := path.Join(sourcePath, zapFolderName)

		for _, fileHash := range fileHashesToUnZap {
			orchestrator.StartTask()
			go ctx.unZapHash(orchestrator, &processedFileIds, zapSourcePath, destinationAbsolutePath, &fileHash)
		}

		orchestrator.WaitForTasks()
	}
}

func (ctx *Context) unZapHash(orchestrator *utils.TaskOrchestrator, processedFileIds *[]uint, zapSourcePath, destinationAbsolutePath string, fileHashToUnZap *ZapResult) {
	destinationPath := path.Join(destinationAbsolutePath, fileHashToUnZap.AbsolutePath)

	// Make the destination directory if required
	err := os.MkdirAll(path.Dir(destinationPath), 0700)

	if err != nil {
		log.Panic(err)
	}

	hexFileName := hex.EncodeToString(base58.Decode(fileHashToUnZap.Hash))

	// un-ZAP
	err = CopyOrMoveFile(path.Join(zapSourcePath, hexFileName), destinationPath, false)

	if err != nil {
		log.Panic(err)
	}

	orchestrator.Lock()
	*processedFileIds = append(*processedFileIds, fileHashToUnZap.ID)
	orchestrator.Unlock()

	orchestrator.FinishTask()
}
