package main

import (
	"data-tools/config"
	"data-tools/utils"
	_ "embed"
	"fmt"
	"log"
)

//go:embed config.yaml
var defaultConfigData []byte

//goland:noinspection GoUnnecessarilyExportedIdentifiers
var AppVersion = "6.0"

var usageText = "Usage: go run main.go [db path]"

func main() {
	print(fmt.Sprintf("data-tools version %s\n", AppVersion))

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

	//err = ctx.AddRootPath("/Users/antfie/Sync")
	//
	//if err != nil && !errors.Is(err, ErrPathAlreadyAdded) {
	//	log.Fatal(err)
	//}
	//
	//err = ctx.Crawl()
	//
	//if err != nil {
	//	log.Fatal(err)
	//}

	err = ctx.HashFiles()

	if err != nil {
		log.Fatal(err)
	}
	//
	//err = ctx.SizeFiles()
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//err = ctx.TypeFiles()
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//err = ctx.Zap("foo_output", true)
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//err = ctx.UnZap("foo_output", "bob")
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
}
