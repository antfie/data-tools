package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

func ClearEmptyFolderNew2(path string) error {
	pathInfo, err := os.Stat(path)

	if err != nil {
		return err
	}

	if !pathInfo.IsDir() {
		return fmt.Errorf("%s is not a folder", path)
	}

	isEmpty, err := clearFolder(path)

	print(isEmpty)

	return err

}

func clearFolder(folderPath string) (int, error) {
	entries, err := os.ReadDir(folderPath)

	if err != nil {
		return 0, err
	}

	fileCount := 0

	for _, entry := range entries {
		thisFilePath := path.Join(folderPath, entry.Name())

		if entry.IsDir() {
			childPathFileCount, clearFolderErr := clearFolder(thisFilePath)

			if clearFolderErr != nil {
				return 0, err
			}

			fileCount += childPathFileCount
		} else {
			// Is the file junk?
			if entry.Name() == ".DS_Store" {
				err = os.Remove(thisFilePath)

				if err != nil {
					return 0, err
				}
			} else {
				fileCount++
			}
		}
	}

	if fileCount == 0 {
		err = os.Remove(folderPath)

		if err != nil {
			return 0, err
		}
	}

	return fileCount, nil
}

func clearEmptyFolderNew(path string) error {
	var pathsWithNoFiles []string
	fileCount := 0
	lastPath := path

	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Ignore the root directory
		if p == path {
			return nil
		}

		if d.IsDir() {
			if fileCount == 0 {
				pathsWithNoFiles = append(pathsWithNoFiles, lastPath)
			}

			fileCount = 0
			lastPath = p
		} else {
			fileCount++
			//return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Also do we want to factor in the ignore list?

	// or do this
	// sortFilePathsByShortest
	// and check in the loop below with a range loop, check if the path is in the prefix of already cleared path?

	// Zip backward through the paths with no files
	for i := len(pathsWithNoFiles) - 1; i >= 0; i-- {
		err = os.RemoveAll(pathsWithNoFiles[i])

		if err != nil {
			return err
		}
	}

	return err
}

func clearEmptyFolders(filePaths []string) error {
	foldersToProcess := getPathsForRMDir(filePaths)

	for _, filePath := range foldersToProcess {
		err := clearEmptyFolder(filePath)

		if err != nil {
			return err
		}
	}

	return nil
}

func RemoveEmptyFolders(path string) (bool, error) {

	entries, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("failed to read directory %s: %v", path, err)
	}

	// Recursively remove empty subdirectories
	isEmpty := true
	for _, entry := range entries {
		if entry.IsDir() {
			subDir := filepath.Join(path, entry.Name())
			// Recursively check and remove empty subdirectories
			removed, err := RemoveEmptyFolders(subDir)
			if err != nil {
				return false, err
			}
			if removed {
				fmt.Printf("Removed empty folder: %s\n", subDir)
			} else {
				isEmpty = false
			}
		} else {
			// If there are files, the directory is not empty
			isEmpty = false
		}
	}

	// If the directory is empty, remove it
	if isEmpty {
		err := os.Remove(path)
		if err != nil {
			return false, fmt.Errorf("failed to remove directory %s: %v", path, err)
		}
		return true, nil
	}

	return false, nil
}

func clearEmptyFolder(filePath string) error {
	for {
		foldersDeleted := 0

		err := filepath.Walk(filePath, func(currentPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				removed, err := removeDirectoryIfEmpty(currentPath)

				if err != nil {
					return err
				}

				if removed {
					foldersDeleted++
				}
			}

			return nil
		})

		if err != nil {
			return err
		}

		if foldersDeleted == 0 {
			return nil
		}
	}
}

func removeDirectoryIfEmpty(currentPath string) (bool, error) {
	entries, err := os.ReadDir(currentPath)

	if err != nil {
		return false, err
	}

	if len(entries) == 0 {
		return true, os.Remove(currentPath)
	}

	return false, nil
}
