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
	"gorm.io/gorm"
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

	// This is needed for the progress bar
	println()

	err = ctx.removeDuplicates(safeMode)

	if err != nil {
		return err
	}

	// This is needed for the progress bar
	println()

	return ctx.removeEmptyZappedFolders(safeMode)
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

	remainingPercentage := (float64(info.TotalFileSize-info.UniqueHashTotalFileSize) / float64(info.TotalFileSize)) * 100
	removalPercentage := (float64(info.UniqueHashTotalFileSize) / float64(info.TotalFileSize)) * 100
	utils.ConsoleAndLogPrintf("Copying %s (%s) to \"%s\". This is %.2f%% of %s, a reduction of %s (%.2f%%)", utils.Pluralize("de-duplicated file", info.FileHashesToZap), humanize.Bytes(info.TotalFileSize-info.UniqueHashTotalFileSize), outputPathAbs, remainingPercentage, humanize.Bytes(info.TotalFileSize), humanize.Bytes(info.TotalFileSize-(info.TotalFileSize-info.UniqueHashTotalFileSize)), removalPercentage)

	bar := progressbar.Default(info.FileHashesToZap)

	zapStructureCreated := false

	// Do batches until there are no more
	for {
		var fileHashesToZap []ZapResult
		result = ctx.DB.Raw(QueryGetFileHashesToZapWithLimit(), ctx.Config.BatchSize).Scan(&fileHashesToZap)

		if result.Error != nil {
			return result.Error
		}

		// Have we finished?
		if len(fileHashesToZap) == 0 {
			return nil
		}

		// We do this here as there is no point in creating a structure if there is nothing to hash (i.e. calling this earlier)
		if !zapStructureCreated {
			err = createZapDirectoryStructure(outputPathAbs)

			if err != nil {
				return err
			}
			zapStructureCreated = true
		}

		var zappedFileHashIds []uint
		var zappedFileIds []uint
		var notFoundFileIDs []uint

		orchestrator := utils.NewTaskOrchestrator(bar, len(fileHashesToZap), ctx.Config.MaxConcurrentFileOperations)

		for _, fileHash := range fileHashesToZap {
			orchestrator.StartTask()
			go ctx.zapFile(orchestrator, safeMode, outputPathAbs, fileHash, &zappedFileHashIds, &zappedFileIds, &notFoundFileIDs)
		}

		orchestrator.WaitForTasks()

		err = ctx.DB.Transaction(func(tx *gorm.DB) error {
			if len(zappedFileHashIds) > 0 {
				result = tx.Where("id IN ?", zappedFileHashIds).Updates(models.FileHash{
					Zapped: true,
				})

				if result.Error != nil {
					return result.Error
				}
			}

			if len(zappedFileIds) > 0 {
				result = tx.Where("id IN ?", zappedFileIds).Updates(models.File{
					Zapped: true,
				})

				if result.Error != nil {
					return result.Error
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
			log.Fatalf("DB Error: %v", err)
		}
	}
}

func (ctx *Context) zapFile(orchestrator *utils.TaskOrchestrator, safeMode bool, zapBasePath string, file ZapResult, zappedFileHashIds, zappedFileIds, notFoundFileIDs *[]uint) {
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
	destinationPath := path.Join(zapBasePath, hexFileName[:2], hexFileName[2:4], hexFileName[4:])
	err := CopyOrMoveFile(file.AbsolutePath, destinationPath, move, true)

	if err != nil {
		log.Fatalf("Could not ZAP file \"%s\": %v", file.AbsolutePath, err)
	}

	orchestrator.Lock()
	*zappedFileHashIds = append(*zappedFileHashIds, file.FileHashID)
	*zappedFileIds = append(*zappedFileIds, file.FileID)
	orchestrator.Unlock()

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
		if len(duplicateFilesToRemove) == 0 {
			return nil
		}

		orchestrator := utils.NewTaskOrchestrator(bar, len(duplicateFilesToRemove), ctx.Config.MaxConcurrentFileOperations)

		var zappedFileIds []uint
		var notFoundFileIDs []uint

		for _, file := range duplicateFilesToRemove {
			orchestrator.StartTask()
			go ctx.removeDuplicateFile(orchestrator, safeMode, file, &zappedFileIds, &notFoundFileIDs)
		}

		orchestrator.WaitForTasks()

		err := ctx.DB.Transaction(func(tx *gorm.DB) error {
			if len(zappedFileIds) > 0 {
				result = tx.Where("id IN ?", zappedFileIds).Updates(models.File{
					Zapped: true,
				})

				if result.Error != nil {
					return result.Error
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
			log.Fatalf("DB Error: %v", err)
		}
	}
}

func (ctx *Context) removeDuplicateFile(orchestrator *utils.TaskOrchestrator, safeMode bool, file FileIdAndPath, zappedFileIds, notFoundFileIDs *[]uint) {
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

		if err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Fatalf("Could not remove file \"%s\": %v", file.AbsolutePath, err)
		}
	}

	orchestrator.Lock()
	*zappedFileIds = append(*zappedFileIds, file.FileID)
	orchestrator.Unlock()

	orchestrator.FinishTask()
}

func (ctx *Context) removeEmptyZappedFolders(safeMode bool) error {
	utils.ConsoleAndLogPrintf("Removing empty folders")

	var filesToProcess []string
	var foldersToProcess []string
	result := ctx.DB.Raw(QueryGetZappedFolders()).Scan(&filesToProcess)

	if result.Error != nil {
		return result.Error
	}

	for _, file := range filesToProcess {
		pathName := path.Dir(file)

		if utils.IsInArray(path.Base(pathName), ctx.Config.FolderNamesToIgnore) {
			continue
		}

		if !utils.IsInArray(pathName, foldersToProcess) {
			foldersToProcess = append(foldersToProcess, pathName)
		}
	}

	if safeMode {
		return nil
	}

	return ClearEmptyFolders(foldersToProcess)
}

// This will create 65,536 'buckets' in which to store the data from 00/00 to ff/ff
func createZapDirectoryStructure(absoluteBasePath string) error {
	info, err := os.Stat(absoluteBasePath)

	// If a folder already exists at this location we assume the Zap structure has already been created
	if info != nil || !errors.Is(err, os.ErrNotExist) {
		return err
	}

	utils.ConsoleAndLogPrintf("Creating ZAP data structure")

	err = osMkdir(absoluteBasePath)

	if err != nil {
		return err
	}

	for level1 := 0; level1 < 0x100; level1++ {
		level1Path := path.Join(absoluteBasePath, fmt.Sprintf("%02x", level1))

		err = osMkdir(level1Path)

		if err != nil {
			return err
		}

		for level2 := 0; level2 < 0x100; level2++ {
			level2Path := path.Join(level1Path, fmt.Sprintf("%02x", level2))

			err = osMkdir(level2Path)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
