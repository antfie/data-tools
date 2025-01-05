package main

import (
	"data-tools/config"
	"data-tools/utils"
	_ "embed"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

//goland:noinspection GoUnnecessarilyExportedIdentifiers
var AppVersion = "6.0"

var usageText = "Usage: ./data-tools command.\nAvailable commands:\n  add_root\n  crawl\n  size\n  hash\n  zap\n  unzap\n"

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

	err = ctx.runCommand(strings.ToLower(os.Args[1]))

	if err != nil {
		utils.ConsoleAndLogPrintf("%v", err)
	}

	duration := math.Round(time.Since(startTime).Seconds())
	formattedDuration := fmt.Sprintf("%.0f second", duration)

	if duration != 1 {
		formattedDuration += "s"
	}

	utils.ConsoleAndLogPrintf("Finished in %s", formattedDuration)
}

func (ctx *Context) runCommand(command string) error {
	switch command {
	case "add_root":
		if len(os.Args) != 3 {
			log.Fatal("Move requires source and destination.")
		}
		return ctx.AddRootPath(os.Args[2])

	case "crawl":
		return ctx.Crawl()

	case "size":
		return ctx.SizeFiles()

	case "type":
		return ctx.TypeFiles()

	case "zap":
		if len(os.Args) != 3 {
			log.Fatal("Move requires source and destination.")
		}

		return ctx.Zap(os.Args[2], false)

	case "unzap":
		if len(os.Args) != 4 {
			log.Fatal("Move requires source and destination.")
		}

		return ctx.UnZap(os.Args[2], os.Args[3])
	}

	return errors.New(fmt.Sprintf("Command \"%s\" not recognised. %s", command, usageText))
}
