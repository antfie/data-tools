package main

import (
	"data-tools/models"
	"data-tools/utils"
	"errors"
	"fmt"
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

func (ctx *Context) Zap(safeMode bool) error {
	err := ctx.moveUniqueFilesToZapFolder(safeMode)

	if err != nil {
		return err
	}

	return nil

	//
	//// This is needed for the progress bar
	//println()
	//
	//err = ctx.removeDuplicates(safeMode)
	//
	//if err != nil {
	//	return err
	//}
	//
	//// This is needed for the progress bar
	//println()
	//
	//return ctx.removeEmptyZappedFolders(safeMode)
}

func (ctx *Context) moveUniqueFilesToZapFolder(safeMode bool) error {
	utils.ConsoleAndLogPrintf("Acquiring data...")
	total, batches, err := ctx.GetBatchesOfIDs(QueryGetFileIdsToZap(), "f")

	if err != nil {
		return err
	}

	if len(batches) == 0 {
		utils.ConsoleAndLogPrintf("No files to ZAP. Have you already hashed?")
		return nil
	}

	outputPathAbs, err := filepath.Abs(ctx.Config.ZapDataPath)

	if err != nil {
		return err
	}

	err = createZapDirectoryStructure(outputPathAbs)

	if err != nil {
		return err
	}

	utils.ConsoleAndLogPrintf("Moving %s to \"%s\" in %s", utils.Pluralize("file", total), ctx.Config.ZapDataPath, utils.Pluralize("batch", int64(len(batches))))

	bar := progressbar.Default(total)

	for _, batch := range batches {
		var fileHashesToZap []ZapResult
		result := ctx.DB.Raw(QueryGetFileHashesToZapMOOO(), batch).Scan(&fileHashesToZap)

		if result.Error != nil {
			return result.Error
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

		transactionErr := ctx.DB.Transaction(func(tx *gorm.DB) error {
			if len(zappedFileHashIds) > 0 {
				result = tx.Where("id IN ?", zappedFileHashIds).Updates(models.FileHash{
					Zapped: true,
				})

				if result.Error != nil {
					return result.Error
				}

				if result.RowsAffected != int64(len(zappedFileHashIds)) {
					return errors.New("could not zap file hashes in db")
				}
			}

			if len(zappedFileIds) > 0 {
				zapFileError := zapFilesInDB(tx, zappedFileIds)

				if zapFileError != nil {
					return zapFileError
				}
			}

			return DealWithNotFoundFiles(tx, notFoundFileIDs)
		})

		if transactionErr != nil {
			return transactionErr
		}
	}

	return nil
}

func zapFilesInDB(tx *gorm.DB, zappedFileIds []uint) error {
	if len(zappedFileIds) > 0 {
		fileUpdateResult := tx.Where("id IN ?", zappedFileIds).Updates(models.File{
			Zapped: true,
		})

		if fileUpdateResult.Error != nil {
			return fileUpdateResult.Error
		}

		if fileUpdateResult.RowsAffected != int64(len(zappedFileIds)) {
			return errors.New("could not zap files in db")
		}
	}

	return nil
}

func DealWithNotFoundFiles(tx *gorm.DB, notFoundFileIDs []uint) error {
	if len(notFoundFileIDs) > 0 {
		notFoundFileResult := tx.Where("id IN ?", notFoundFileIDs).Delete(&models.File{})

		if notFoundFileResult.Error != nil {
			return notFoundFileResult.Error
		}

		if notFoundFileResult.RowsAffected != int64(len(notFoundFileIDs)) {
			return errors.New("could not delete files from db")
		}
	}

	return nil
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

	// Store as hex so this will work fine on case-insensitive filesystems
	hexFileName := DecodeHash(file.Hash)
	destinationPath := path.Join(zapBasePath, FormatRelativeZapFilePathFromHash(hexFileName))

	// Only move if not in safe mode
	move := !safeMode

	// ZAP
	success, err := CopyOrMoveFile(file.AbsolutePath, destinationPath, move, true)

	if err != nil {
		log.Fatalf("Could not ZAP file \"%s\": %v", file.AbsolutePath, err)
	}

	if success {
		orchestrator.Lock()
		*zappedFileHashIds = append(*zappedFileHashIds, file.FileHashID)
		*zappedFileIds = append(*zappedFileIds, file.FileID)
		orchestrator.Unlock()
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
			zapFileError := zapFilesInDB(tx, zappedFileIds)

			if zapFileError != nil {
				return zapFileError
			}

			return DealWithNotFoundFiles(tx, notFoundFileIDs)
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

		if err != nil && !os.IsNotExist(err) {
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
	if info != nil || !os.IsNotExist(err) {
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
