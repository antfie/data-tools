package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func IsDir(path string) bool {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return true
	}

	return false
}

func IsFile(path string) bool {
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return true
	}

	return false
}

func GetTypeOfFile(file string) (string, error) {
	command := exec.Command("file", "-b", "--mime-type", file)
	output, err := command.Output()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func getPathsForMkdirs(filePaths []string) []string {
	var resolvedPaths []string

	sort.Slice(filePaths, func(i, j int) bool {
		return len(filePaths[i]) > len(filePaths[j])
	})

	for _, filePath := range filePaths {
		basePath := filepath.Dir(filePath)
		found := false
		for _, existingPath := range resolvedPaths {
			if strings.HasPrefix(existingPath, basePath) {
				found = true
				break
			}
		}

		if !found {
			resolvedPaths = append(resolvedPaths, basePath)
		}
	}

	return resolvedPaths
}
