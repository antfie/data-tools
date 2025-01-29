package main

import (
	"errors"
	"fmt"
	"os"
	"path"
)

func ClearEmptyFolders(path []string) error {
	for _, p := range path {
		pathInfo, err := os.Stat(p)

		if err != nil {
			// Ignore if the file/path was not found
			if os.IsNotExist(err) {
				return nil
			}

			return err
		}

		if !pathInfo.IsDir() {
			return fmt.Errorf("%s is not a folder", path)
		}

		_, err = clearFoldersInternal(p)

		if err != nil {
			return err
		}
	}

	return nil
}

func clearFoldersInternal(folderPath string) (int, error) {
	entries, err := os.ReadDir(folderPath)

	if err != nil {
		return 0, err
	}

	fileCount := 0

	for _, entry := range entries {
		thisFilePath := path.Join(folderPath, entry.Name())

		if entry.IsDir() {
			childPathFileCount, clearFolderErr := clearFoldersInternal(thisFilePath)

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
