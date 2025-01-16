package main

import (
	"data-tools/utils"
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

// file -b --mime-type test.db
// application/vnd.sqlite3

func GetFileType(filePath string) (string, error) {
	command := exec.Command("file", "-b", "--mime-type", filePath)
	output, err := command.Output()

	if err != nil {
		return "", err
	}

	formatted := string(output)

	// Deal with this type of scenario
	// file -b --mime-type "/some/path/Widgets/HP Ink Widget.wdgt/HPInkWidget.widgetplugin/Contents/MacOS/InkWidgetTool"
	// application/x-mach-binary
	// /some/path/Widgets/HP Ink Widget.wdgt/HPInkWidget.widgetplugin/Contents/MacOS/InkWidgetTool (for architecture ppc):	application/x-mach-binary
	// /some/path/Widgets/HP Ink Widget.wdgt/HPInkWidget.widgetplugin/Contents/MacOS/InkWidgetTool (for architecture i386):	application/x-mach-binary
	if strings.Contains(formatted, "\n") {
		formatted = strings.Split(formatted, "\n")[0]
	}

	formatted = strings.TrimSpace(formatted)

	return formatted, nil
}

func sortFilePathsByLongest(filePaths []string) {
	sort.Slice(filePaths, func(i, j int) bool {
		return strings.Count(filePaths[i], string(os.PathSeparator)) > strings.Count(filePaths[j], string(os.PathSeparator))
	})
}

func sortFilePathsByShortest(filePaths []string) {
	sort.Slice(filePaths, func(i, j int) bool {
		return strings.Count(filePaths[i], string(os.PathSeparator)) < strings.Count(filePaths[j], string(os.PathSeparator))
	})
}

func getPathsForMkdirs(filePaths []string) []string {
	var resolvedPaths []string

	sortFilePathsByLongest(filePaths)

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

func getPathsForRMDir(filePaths []string) []string {
	var resolvedPaths []string

	sortFilePathsByShortest(filePaths)

	for _, filePath := range filePaths {
		basePath := filepath.Dir(filePath)
		found := false
		for _, existingPath := range resolvedPaths {
			if !strings.HasPrefix(existingPath, basePath) {
				found = true
				break
			}
		}

		if !found {
			if !utils.IsInArray(basePath, resolvedPaths) {
				resolvedPaths = append(resolvedPaths, basePath)
			}
		}
	}

	return resolvedPaths
}

func GetAllFiles(rootPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootPath, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			absoluteFileName, err := filepath.Abs(currentPath)

			if err != nil {
				return err
			}

			files = append(files, absoluteFileName)
		}

		return nil
	})

	return files, err
}

func osMkdir(path string) error {
	return os.Mkdir(path, 0700)
}

func osMkdirAll(path string) error {
	return os.MkdirAll(path, 0700)
}
