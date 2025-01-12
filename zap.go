package main

import (
	"data-tools/models"
	"data-tools/utils"
	"encoding/hex"
	"errors"
	"fmt"
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
	err := ctx.copyDeduplicatedFiles(outputPath, safeMode)

	if err != nil {
		return err
	}

	//err = ctx.removeDuplicates(safeMode)

	if err != nil {
		return err
	}

	//return ctx.removeEmptyZappedFolders(safeMode)
	return err
}

func (ctx *Context) copyDeduplicatedFiles(outputPath string, safeMode bool) error {
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

	err = createZapDirectoryStructure(outputPathAbs)

	if err != nil {
		return err
	}

	remainingPercentage := (float64(info.TotalFileSize-info.UniqueHashTotalFileSize) / float64(info.TotalFileSize)) * 100
	removalPercentage := (float64(info.UniqueHashTotalFileSize) / float64(info.TotalFileSize)) * 100
	utils.ConsoleAndLogPrintf("Copying %s (%s) to \"%s\". This is %.2f%% of %s, a reduction of %s (%.2f%%)", utils.Pluralize("de-duplicated file", info.FileHashesToZap), humanize.Bytes(info.TotalFileSize-info.UniqueHashTotalFileSize), outputPathAbs, remainingPercentage, humanize.Bytes(info.TotalFileSize), humanize.Bytes(info.TotalFileSize-(info.TotalFileSize-info.UniqueHashTotalFileSize)), removalPercentage)

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
			return nil
		}

		var notFoundFileIDs []uint

		orchestrator := utils.NewTaskOrchestrator(bar, len(fileHashesToZap), ctx.Config.MaxConcurrentFileOperations)

		for _, fileHash := range fileHashesToZap {
			orchestrator.StartTask()
			go ctx.zapFile(orchestrator, safeMode, outputPathAbs, fileHash, &notFoundFileIDs)
		}

		orchestrator.WaitForTasks()
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

	// TODO: for collision detection, no do not compare the files. If file with same hash then consider it the same, no need to re-hash, just ignore, move on
	err := CopyOrMoveFile(file.AbsolutePath, path.Join(zapBasePath, hexFileName[:2], hexFileName[2:4], hexFileName[4:]), move)

	if err != nil {
		log.Fatalf("Could not ZAP file \"%s\": %v", file.AbsolutePath, err)
	}

	result := ctx.DB.Model(&models.FileHash{}).Where("id = ?", file.FileHashID).Updates(models.FileHash{
		Zapped: true,
	})

	if result.Error != nil {
		log.Fatalf("DB Error: %v", result.Error)
	}

	result = ctx.DB.Model(&models.File{}).Where("id = ?", file.FileID).Updates(models.File{
		Zapped: true,
	})

	if result.Error != nil {
		log.Fatalf("DB Error: %v", result.Error)
	}

	orchestrator.FinishTask()
}

func (ctx *Context) removeDuplicates(safeMode bool) error {
	type ZapInfo struct {
		CountOfFilesToRemove  int64
		TotalFileSizeToRemove uint64
	}

	var info ZapInfo
	result := ctx.DB.Raw(`
SELECT		COUNT(*) count_of_files_to_remove,
			SUM(f.size) total_file_size_to_remove
FROM 		files f
JOIN 		file_hashes fh ON f.file_hash_id = fh.id
WHERE		f.zapped = 0
AND			f.deleted_at IS NULL
AND			f.ignored = 0
AND			fh.zapped = 1
AND			fh.ignored = 0
`).First(&info)

	utils.ConsoleAndLogPrintf("Removing %s (%s)", utils.Pluralize("duplicate file", info.CountOfFilesToRemove), humanize.Bytes(info.TotalFileSizeToRemove))

	bar := progressbar.Default(info.CountOfFilesToRemove)

	// Do batches until there are no more
	for {
		var duplicateFilesToRemove []FileIdAndPath
		result = ctx.DB.Raw(QueryGetDuplicateFilesToZapWithLimit(), ctx.Config.BatchSize).Scan(&duplicateFilesToRemove)

		if result.Error != nil {
			return result.Error
		}

		// Have we finished?
		if duplicateFilesToRemove == nil {
			return nil
		}

		orchestrator := utils.NewTaskOrchestrator(bar, len(duplicateFilesToRemove), ctx.Config.MaxConcurrentFileOperations)

		var notFoundFileIDs []uint

		for _, file := range duplicateFilesToRemove {
			orchestrator.StartTask()
			go ctx.removeDuplicateFile(orchestrator, safeMode, file, &notFoundFileIDs)
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

func (ctx *Context) removeDuplicateFile(orchestrator *utils.TaskOrchestrator, safeMode bool, file FileIdAndPath, notFoundFileIDs *[]uint) {
	// If the file does not exist we can ignore it
	if !IsFile(file.AbsolutePath) {
		orchestrator.Lock()
		log.Printf("Ignoring not-found file \"%s\"", file.AbsolutePath)
		*notFoundFileIDs = append(*notFoundFileIDs, file.FileID)
		orchestrator.Unlock()

		orchestrator.FinishTask()
		return
	}

	if !safeMode {
		err := os.Remove(file.AbsolutePath)

		if err != nil {
			log.Fatalf("Could not remove file \"%s\": %v", file.AbsolutePath, err)
		}
	}

	result := ctx.DB.Model(&models.File{}).Where("id = ?", file.FileID).Updates(models.File{
		Zapped: true,
	})

	if result.Error != nil {
		log.Fatalf("DB Error: %v", result.Error)
	}

	orchestrator.FinishTask()
}

func (ctx *Context) removeEmptyZappedFolders(safeMode bool) error {
	utils.ConsoleAndLogPrintf("Removing empty folders")

	var foldersToProcess []string
	result := ctx.DB.Raw(QueryGetZappedFolders()).Scan(&foldersToProcess)

	if result.Error != nil {
		return result.Error
	}

	if safeMode {
		return nil
	}

	return clearEmptyFolders(foldersToProcess)
}

func createZapDirectoryStructure(absoluteBasePath string) error {
	info, err := os.Stat(absoluteBasePath)

	if info != nil || !errors.Is(err, os.ErrNotExist) {
		return err
	}

	utils.ConsoleAndLogPrintf("Creating ZAP data structure")

	err = os.Mkdir(absoluteBasePath, 0700)

	if err != nil {
		return err
	}

	for x := 0; x < 0x100; x++ {
		pathX := path.Join(absoluteBasePath, fmt.Sprintf("%02x", x))

		err = os.Mkdir(pathX, 0700)

		if err != nil {
			return err
		}

		for y := 0; y < 0x100; y++ {
			pathY := path.Join(pathX, fmt.Sprintf("%02x", y))

			err = os.Mkdir(pathY, 0700)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
