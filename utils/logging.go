package utils

import (
	"fmt"
	"log"
	"os"
	"path"
)

func SetupLogger(filePath string) error {
	logFile, err := os.OpenFile(path.Clean(filePath), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)

	if err != nil {
		return err
	}

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.LUTC)
	return nil
}

func ConsoleAndLogPrintf(format string, v ...interface{}) {
	log.Printf(format, v...)
	print(fmt.Sprintf(format, v...) + "\n")
}
