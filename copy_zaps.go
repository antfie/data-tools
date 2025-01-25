package main

import (
	"data-tools/utils"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"os"

	"path"
)

func MergeZaps(sourcePath, destinationPath string) error {
	sourcePathInfo, err := os.Stat(sourcePath)

	if err != nil {
		return err
	}

	if !sourcePathInfo.IsDir() {
		return fmt.Errorf("\"%s\" is not a directory", sourcePath)
	}

	destinationPathInfo, err := os.Stat(destinationPath)

	if err != nil {
		return err
	}

	if !destinationPathInfo.IsDir() {
		return fmt.Errorf("\"%s\" is not a directory", destinationPathInfo)
	}

	print(fmt.Printf("This will move zaps from :\"%s\" to \"%s\". If you wish to proceed type YES: ", sourcePath, destinationPath))
	var input string
	_, err = fmt.Scanln(&input)

	if err != nil {
		return err
	}

	if input != "YES" {
		return nil
	}

	paths := buildPathMap(sourcePath, destinationPath)
	bar := progressbar.Default(int64(len(paths)))
	orchestrator := utils.NewTaskOrchestrator(bar, len(paths), 10)

	for sourceFilePath, destinationFilePath := range paths {
		orchestrator.StartTask()
		go copyZapsInFolder(orchestrator, sourceFilePath, destinationFilePath)
	}

	orchestrator.WaitForTasks()

	return nil
}

func buildPathMap(sourcePath, destinationPath string) map[string]string {
	paths := map[string]string{}

	for x := 0; x < 0x100; x++ {
		for y := 0; y < 0x100; y++ {
			src := path.Join(sourcePath, fmt.Sprintf("%02x", x), fmt.Sprintf("%02x", y))
			dst := path.Join(destinationPath, fmt.Sprintf("%02x", x), fmt.Sprintf("%02x", y))
			paths[src] = dst
		}
	}

	return paths
}

func copyZapsInFolder(orchestrator *utils.TaskOrchestrator, sourcePath, destinationPath string) {
	err := CopyOrMoveFiles(sourcePath, destinationPath, true, true)

	if err != nil {
		//return err
	}

	orchestrator.FinishTask()
}
