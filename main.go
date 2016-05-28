package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/getblank/blank-filestore/intranet"
	_ "github.com/getblank/blank-filestore/store"
)

func main() {
	log.SetLevel(log.DebugLevel)
	var srAddress *string
	rootCmd := &cobra.Command{
		Use:   "blank-filestore",
		Short: "File storage microservice for Blank platform",
		Long:  "The next generation of web applications. This is the file storage subsystem.",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("blank-filestore started")
			intranet.Init(*srAddress)
		},
	}

	srAddress = rootCmd.PersistentFlags().StringP("service-registry", "s", "ws://localhost:1234", "Service registry uri")

	if err := rootCmd.Execute(); err != nil {
		println(err.Error())
		os.Exit(-1)
	}
}
