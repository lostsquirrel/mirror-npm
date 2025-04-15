package main

import (
	"flag"
	"fmt"
	"mirror-npm/app"
	"mirror-npm/utils"
	"os"
)

func init() {
	utils.LoadEnv()
}

func main() {
	server := flag.Bool("s", false, "Run as server")
	help := flag.Bool("h", false, "Show help")
	flag.Parse()

	switch {
	case *help:
		fmt.Println("Usage: mirror-npm [-s]")
		os.Exit(0)
	case *server:
		instance := app.NewInstance()
		instance.Start()
	default:
		utils.UpdateMetaOnDisk()
	}
}
