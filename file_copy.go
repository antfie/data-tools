package main

import (
	"io"
	"os"
	"path"
	"path/filepath"
)

func CopyOrMoveFile(source, destination string, move bool) error {
	destinationIsTheSame, err := isDestinationTheSame(source, destination)

	if err != nil {
		return err
	}

	if destinationIsTheSame {
		if move {
			// Remove existing file
			return os.Remove(source)
		}

		// Nothing to do?
		return nil
	}

	if move {
		return os.Rename(source, destination)
	}

	sourceFile, err := os.Open(path.Clean(source))

	if err != nil {
		return err
	}

	// Create the directory structure if required
	err = os.MkdirAll(filepath.Dir(destination), 0750)

	if err != nil {
		return err
	}

	destinationFile, err := os.Create(path.Clean(destination))

	if err != nil {
		return err
	}

	_, err = io.Copy(destinationFile, sourceFile)

	if err != nil {
		return err
	}

	err = destinationFile.Sync()

	if err != nil {
		return err
	}

	err = destinationFile.Close()

	if err != nil {
		return err
	}

	err = sourceFile.Close()

	if err != nil {
		return err
	}

	return nil
}

func isDestinationTheSame(source, destination string) (bool, error) {
	_, err := os.Stat(destination)

	// Does a file already exist at the destination?
	if err == nil {
		filesAreTheSame, err := CompareFiles(source, destination)

		if err != nil {
			return false, err
		}

		// Nothing to do here
		if filesAreTheSame {
			return filesAreTheSame, err
		}

		// The files are different. This is a problem
		return true, ErrNotOverwritingExistingDifferentFile
	}

	// Was the error anything other than an expected file not found?
	if os.IsNotExist(err) {
		return false, nil
	}

	// Was the error anything other than an expected file not found?
	return false, err
}
