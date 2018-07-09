package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/gemnasium/logrus-graylog-hook.v2"

	"github.com/getblank/blank-fs/intranet"
)

var (
	buildTime string
	gitHash   string
	version   = "0.0.33"
)

func main() {
	if os.Getenv("BLANK_DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(&log.JSONFormatter{})
	if os.Getenv("GRAYLOG2_HOST") != "" {
		host := os.Getenv("GRAYLOG2_HOST")
		port := os.Getenv("GRAYLOG2_PORT")
		if port == "" {
			port = "12201"
		}
		source := os.Getenv("GRAYLOG2_SOURCE")
		if source == "" {
			source = "blank-fs"
		}
		hook := graylog.NewGraylogHook(host+":"+port, map[string]interface{}{"source-app": source})
		log.AddHook(hook)
	}

	srAddress := flag.String("s", "ws://localhost:1234", "Service registry uri")
	port := flag.String("p", "8082", "TCP port to listen")
	verFlag := flag.Bool("v", false, "Prints version and exit")
	flag.Parse()

	if *verFlag {
		printVersion()
		return
	}

	if sr := os.Getenv("BLANK_SERVICE_REGISTRY_URI"); len(sr) > 0 {
		srAddress = &sr
	}
	if srPort := os.Getenv("BLANK_SERVICE_REGISTRY_PORT"); len(srPort) > 0 {
		addr := "ws://localhost:" + srPort
		srAddress = &addr
	}
	if fsPort := os.Getenv("BLANK_FILE_STORE_PORT"); len(fsPort) > 0 {
		port = &fsPort
	}

	log.Info("blank-fs started")
	intranet.Init(*srAddress, *port)
}

func printVersion() {
	fmt.Printf("blank-fs: \tv%s \t build time: %s \t commit hash: %s \n", version, buildTime, gitHash)
}
