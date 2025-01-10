package main

import (
	"data-tools/config"
	"data-tools/utils"
	_ "embed"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"log"
	"os"
	"strings"
	"time"
)

//goland:noinspection GoUnnecessarilyExportedIdentifiers
var AppVersion = "6.0"

var usageText = "Usage: ./data-tools command.\nAvailable commands:\n  crawl\n  hash\n  zap\n  unzap\n"

//go:embed config.yaml
var defaultConfigData []byte

func main() {
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

	err = ctx.runCommand(strings.ToLower(os.Args[1]))

	if err != nil {
		utils.ConsoleAndLogPrintf("Error: %v", err)
	}

	utils.ConsoleAndLogPrintf("Finished in %s", utils.FormatDuration(time.Since(startTime)))
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
		if len(os.Args) != 3 {
			log.Fatal("zap requires a destination path.")
		}

		return ctx.Zap(os.Args[2], false)

	case "unzap":
		if len(os.Args) != 4 {
			log.Fatal("unzap requires a source and destination path.")
		}

		return ctx.UnZap(os.Args[2], os.Args[3])
	}

	return errors.New(fmt.Sprintf("Command \"%s\" not recognised. %s", command, usageText))
}
