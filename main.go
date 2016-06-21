package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/getblank/blank-fs/intranet"
)

func main() {
	log.SetLevel(log.DebugLevel)
	var srAddress *string
	var port *string
	rootCmd := &cobra.Command{
		Use:   "blank-fs",
		Short: "File storage microservice for Blank platform",
		Long:  "The next generation of web applications. This is the file storage subsystem.",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("blank-fs started")
			intranet.Init(*srAddress, *port)
		},
	}

	srAddress = rootCmd.PersistentFlags().StringP("service-registry", "s", "ws://localhost:1234", "Service registry uri")
	port = rootCmd.PersistentFlags().StringP("port", "p", "8082", "TCP port to listen")

	if err := rootCmd.Execute(); err != nil {
		println(err.Error())
		os.Exit(-1)
	}
}
