package main

import (
	"data-tools/config"
	"data-tools/utils"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
)

//go:embed config.yaml
var defaultConfigData []byte

//goland:noinspection GoUnnecessarilyExportedIdentifiers
var AppVersion = "0.0"

var usageText = "Usage: go run main.go [db path]"

func main() {
	print(fmt.Sprintf("Data Tools version %s\n", AppVersion))

	c, err := config.Load(defaultConfigData)

	if err != nil {
		log.Fatal(err)
	}

	err = utils.SetupLogger(c.LogFilePath)

	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) < 2 {
		log.Fatal("No DB path specified. " + usageText)
	}

	db := initDb(c, os.Args[1])

	ctx := &Context{
		Config: c,
		DB:     db,
	}

	err = ctx.AddRootPath("/Users/antfie/Sync")

	if err != nil && !errors.Is(err, ErrPathAlreadyAdded) {
		log.Fatal(err)
	}

	err = ctx.Crawl()

	if err != nil {
		log.Fatal(err)
	}

	err = ctx.HashFiles()

	if err != nil {
		log.Fatal(err)
	}

	err = ctx.SizeFiles()

	if err != nil {
		log.Fatal(err)
	}

	err = ctx.TypeFiles()

	if err != nil {
		log.Fatal(err)
	}

	err = ctx.Zap("foo_output", true)

	if err != nil {
		log.Fatal(err)
	}

	err = ctx.UnZap("foo_output", "bob")

	if err != nil {
		log.Fatal(err)
	}
}
