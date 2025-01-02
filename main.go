package main

import (
	"data-tools/config"
	"data-tools/utils"
	_ "embed"
	"fmt"
	"github.com/dustin/go-humanize"
	"log"
	"math"
	"time"
)

//go:embed config.yaml
var defaultConfigData []byte

//goland:noinspection GoUnnecessarilyExportedIdentifiers
var AppVersion = "6.0"

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

	err = ctx.AddRootPath("/Users/antfie/Sync")
	//
	//if err != nil && !errors.Is(err, ErrPathAlreadyAdded) {
	//	log.Fatal(err)
	//}
	//
	err = ctx.Crawl()

	//if err != nil {
	//	log.Fatal(err)
	//}

	err = ctx.HashFiles()

	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	err = ctx.SizeFiles()
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	err = ctx.TypeFiles()
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	err = ctx.Zap("foo_output", true)
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	err = ctx.UnZap("foo_output", "bob")
	//

	if err != nil {
		log.Fatal(err)
	}

	duration := math.Round(time.Since(startTime).Seconds())
	formattedDuration := fmt.Sprintf("%.0f second", duration)

	if duration != 1 {
		formattedDuration += "s"
	}

	utils.ConsoleAndLogPrintf("Finished in %s", formattedDuration)
}
