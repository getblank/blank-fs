package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/getblank/blank-fs/intranet"
)

var (
	buildTime string
	gitHash   string
	version   = "0.0.36"
)

func main() {
	if os.Getenv("BLANK_DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(&log.JSONFormatter{})

	port := flag.String("p", "8082", "TCP port to listen")
	verFlag := flag.Bool("v", false, "Prints version and exit")
	flag.Parse()

	if *verFlag {
		printVersion()
		return
	}

	if fsPort := os.Getenv("BLANK_FILE_STORE_PORT"); len(fsPort) > 0 {
		port = &fsPort
	}

	log.Info("blank-fs started")
	intranet.Init(*port)
}

func printVersion() {
	fmt.Printf("blank-fs: \tv%s \t build time: %s \t commit hash: %s \n", version, buildTime, gitHash)
}
