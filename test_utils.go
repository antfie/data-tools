package main

import (
	"log"
	"os"
	"path"
)

var testDataPath = path.Join("test", "data")

func createTempTestDataPath() string {
	tempTestDataPath, err := os.MkdirTemp("", "data-tools-")

	if err != nil {
		log.Fatal(err)
	}

	return tempTestDataPath
}
