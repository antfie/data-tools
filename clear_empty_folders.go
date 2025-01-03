package main

import (
	"os"
	"path/filepath"
)

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
