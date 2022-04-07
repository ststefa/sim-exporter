package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "sim-exporter",
		Short: "Export synthetic prometheus metrics",
		Long:  "Produce synthetic metrics usable as a mock for prometheusscrape-testing.",
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config_file", configFile, "Configuration file to use")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", debug, "Enable debug output")
	// https://le-gall.bzh/post/go/integrating-logrus-with-cobra/
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if debug == true {
			log.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
}
