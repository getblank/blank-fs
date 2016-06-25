package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/getblank/blank-fs/intranet"
)

var (
	buildTime string
	gitHash   string
	version   = "0.0.11"
)

func main() {
	if os.Getenv("BLANK_DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}
	var srAddress *string
	var port *string
	var verFlag *bool
	rootCmd := &cobra.Command{
		Use:   "blank-fs",
		Short: "File storage microservice for Blank platform",
		Long:  "The next generation of web applications. This is the file storage subsystem.",
		Run: func(cmd *cobra.Command, args []string) {
			if *verFlag {
				printVersion()
				return
			}
			log.Info("blank-fs started")
			intranet.Init(*srAddress, *port)
		},
	}

	srAddress = rootCmd.PersistentFlags().StringP("service-registry", "s", "ws://localhost:1234", "Service registry uri")
	port = rootCmd.PersistentFlags().StringP("port", "p", "8082", "TCP port to listen")
	verFlag = rootCmd.PersistentFlags().BoolP("version", "v", false, "Prints version and exit")

	if err := rootCmd.Execute(); err != nil {
		println(err.Error())
		os.Exit(-1)
	}
}

func printVersion() {
	fmt.Printf("blank-fs: \tv%s \t build time: %s \t commit hash: %s \n", version, buildTime, gitHash)
}
