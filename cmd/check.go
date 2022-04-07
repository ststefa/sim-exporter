package cmd

import (
	"container/list"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkCmd)
}

func check(cmd *cobra.Command, args []string) {
	config, err := readAndValidateConfig(configFile)
	if err != nil {
		log.Fatalf("Config problem: %v", err)
	}

	log.Debugf("config:\n%v\n", &config)

	metricNames := list.New()
	for _, value := range config.Metrics {
		metricNames.PushFront(value.Name)
	}
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "check simulation configuration",
	Long:  "Check the metric simulation configuration file.",
	Run:   check,
}
