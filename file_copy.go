package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func CopyOrMoveFiles(source, destination string, move, isDestinationZap bool) error {
	files, err := GetAllFiles(source)

	if err != nil {
		return err
	}

	copyText := "copying"

	if move {
		copyText = "moving"
	}

	for _, sourceFilePath := range files {
		relativePath := strings.TrimPrefix(sourceFilePath, source)
		destinationFilePath := path.Join(destination, relativePath)

		// If there is an error, log it and move onto the next file
		_, err = CopyOrMoveFile(sourceFilePath, destinationFilePath, move, isDestinationZap)

		if err != nil {
			log.Printf("Error %s file \"%s\" to \"%s\": %s\n", copyText, sourceFilePath, destinationFilePath, err)
		}
	}

	return nil
}

func CopyOrMoveFile(source, destination string, move, isDestinationZap bool) (bool, error) {
	comparisonMode := Hash

	// When working with ZAP files for performance we only compare by size as the file name is the hash
	if isDestinationZap {
		comparisonMode = Size
	}

	comparisonResult, comparisonError := isDestinationTheSame(source, destination, comparisonMode)

	if comparisonError != nil {
		return false, comparisonError
	}

	if comparisonResult == Different {
		copyText := "copying"

		if move {
			copyText = "moving"
		}

		log.Printf("Not %s file \"%s\" to \"%s\" because they are different\n", copyText, source, destination)
		return false, nil
	}

	if comparisonResult == Same {
		if move {
			// Remove existing file
			removeErr := os.Remove(source)
			return removeErr == nil, removeErr
		}

		// Nothing to do
		return true, nil
	}

	if comparisonResult == DestinationDoesNotExist {
		// If not working with Zap
		if !isDestinationZap {
			// Create the directory structure if required
			osMkdirAllErr := osMkdirAll(filepath.Dir(destination))

			if osMkdirAllErr != nil {
				return false, osMkdirAllErr
			}
		}

		if move {
			osMoveErr := osMove(source, destination)
			return osMoveErr == nil, osMoveErr
		}

		osCopyErr := osCopy(source, destination)
		return osCopyErr == nil, osCopyErr
	}

	return false, errors.New("comparisonResult test not implemented")
}

// we use the OS rather than golang API to get around limitations e.g. file operations across different filesystems
func osMove(source, destination string) error {
	command := exec.Command("/bin/mv", source, destination)
	return debuggableExecution(command)
}

// we use the OS rather than golang API to get around imitations e.g. file operations across different filesystems
func osCopy(source, destination string) error {
	command := exec.Command("/bin/cp", source, destination)
	return debuggableExecution(command)
}

func debuggableExecution(cmd *exec.Cmd) error {
	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}

type ComparisonMode int

const (
	Size ComparisonMode = iota
	Hash
)

type ShouldOverWrite int

const (
	Indeterminate ShouldOverWrite = iota
	DestinationDoesNotExist
	Same
	Different
)

func isDestinationTheSame(source, destination string, mode ComparisonMode) (ShouldOverWrite, error) {
	destInf, destinationStatErr := os.Stat(destination)

	// Does a file already exist at the destination?
	if os.IsNotExist(destinationStatErr) {
		return DestinationDoesNotExist, nil
	}

	if destinationStatErr != nil {
		return Indeterminate, destinationStatErr
	}

	srcInf, sourceStatErr := os.Stat(source)

	// If we can't find the source file, that would be a problem
	if sourceStatErr != nil {
		return Indeterminate, sourceStatErr
	}

	if mode == Size {
		filesAreTheSame := destInf.Size() == srcInf.Size()

		if filesAreTheSame {
			return Same, nil
		}

		return Different, nil
	}

	if mode == Hash {
		filesAreTheSame, err := CompareFiles(source, destination)

		if err != nil {
			return Indeterminate, err
		}

		// Nothing to do here
		if filesAreTheSame {
			return Same, nil
		}

		return Different, nil
	}

	return Indeterminate, errors.New("ComparisonMode not implemented")
}
