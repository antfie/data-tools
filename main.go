package main

import (
	"data-tools/config"
	"data-tools/crypto"
	"data-tools/utils"
	_ "embed"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//goland:noinspection GoUnnecessarilyExportedIdentifiers
var AppVersion = "6.0"

var usageText = "Usage: ./data-tools command.\nAvailable commands:\n  crawl\n  hash\n  zap\n  unzap\n  merge_zaps\n  clear_empty_folders\n  integrity\n  hash_file\n"

//go:embed config.yaml
var defaultConfigData []byte

func main() {
	sanityCheckOSRequirements()

	c, err := config.Load(defaultConfigData)

	if err != nil {
		log.Fatal(err)
	}

	err = utils.SetupLogger(c.LogFilePath)

	if err != nil {
		log.Fatal(err)
	}

	ctx := &Context{
		Config: c,
		DB:     initDb(c),
	}

	debugFormat := ""

	if c.IsDebug {
		debugFormat = " (debug)"
	}

	utils.ConsoleAndLogPrintf("Data Tools version %s%s. Using %s for file operations and batches of %s", AppVersion, debugFormat, utils.Pluralize("thread", ctx.Config.MaxConcurrentFileOperations), humanize.Comma(ctx.Config.BatchSize))
	startTime := time.Now()

	if len(os.Args) < 2 {
		utils.ConsoleAndLogPrintf(fmt.Sprintf("A command must be specified. %s", usageText))
		return
	}

	command := os.Args[1]
	utils.ConsoleAndLogPrintf("Running command: %s", command)

	err = ctx.runCommand(strings.ToLower(command))

	if err != nil {
		utils.ConsoleAndLogPrintf("Error: %v", err)
	}

	utils.ConsoleAndLogPrintf("Finished in %s", utils.FormatDuration(time.Since(startTime)))
}

func sanityCheckOSRequirements() {
	requiredPrograms := []string{
		"/bin/mv",
		"/bin/cp",
		"/usr/bin/file",
	}

	for _, requiredProgram := range requiredPrograms {
		_, err := os.Stat(requiredProgram)

		if os.IsNotExist(err) {
			log.Fatalf("Error: Could not find required \"%s\" executable", requiredProgram)
		}
	}
}

func (ctx *Context) runCommand(command string) error {
	switch command {
	case "crawl":
		if len(os.Args) != 3 {
			log.Fatal("add_root requires a root path.")
		}

		return ctx.Crawl(os.Args[2])

	case "hash":
		return ctx.HashFiles()

	case "zap":
		return ctx.Zap(false)

	case "unzap":
		if len(os.Args) != 4 {
			log.Fatal("unzap requires source and destination paths.")
		}

		return ctx.UnZap(os.Args[2], os.Args[3])

	case "merge_zaps":
		if len(os.Args) != 4 {
			log.Fatal("merge_zaps requires source and destination paths.")
		}

		return MergeZaps(os.Args[2], os.Args[3])

	case "clear_empty_folders":
		if len(os.Args) != 3 {
			log.Fatal("clear_empty_folders requires a path.")
		}

		return ClearEmptyFolders([]string{os.Args[2]})

	case "integrity":
		return ctx.ZapDBIntegrityTestBySize()

	case "hash_file":
		if len(os.Args) != 3 {
			log.Fatal("hash_file requires a file path.")
		}

		filePath, err := filepath.Abs(os.Args[2])

		if err != nil {
			return err
		}

		utils.ConsoleAndLogPrintf("Hashing \"%s\"", filePath)

		hash, err := crypto.HashFile(filePath)

		if err != nil {
			utils.ConsoleAndLogPrintf("Error: Could not hash file \"%s\": %v", filePath, err)
			return err
		}

		utils.ConsoleAndLogPrintf("Hash of \"%s\" is %s (%s)", filePath, hash, DecodeHash(hash))
		return nil
	}

	return errors.New(fmt.Sprintf("Command \"%s\" not recognised. %s", command, usageText))
}
