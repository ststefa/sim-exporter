package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Whether to enable debug output
	debug = false

	rootCmd = &cobra.Command{
		Use:   "sim-exporter",
		Short: "Export synthetic prometheus metrics",
		Long:  "Produce synthetic metrics usable as a mock for prometheus scrape-testing.",
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", debug, "Enable debug output")
	// https://le-gall.bzh/post/go/integrating-logrus-with-cobra/
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if debug {
			log.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
