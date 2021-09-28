package main

import (
	"log"
	"mirror-npm/app"
	"mirror-npm/utils"
	"os"
)

func init() {
	utils.LoadEnv()
}

const serverFlag = "-s"
const clientFlag = "-c"

func main() {
	if len(os.Args) == 1 || serverFlag == os.Args[1] {
		server := app.NewInstance()
		server.Start()
	} else if len(os.Args) > 1 && clientFlag == os.Args[1] {
		utils.UpdateMetaOnDisk()
	} else {
		log.Fatalf("unknown flag %v", os.Args)
	}

}
