package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var checkCmd = &cobra.Command{
	Use:   "check [simulation-file.yaml]",
	Short: "Validate simulation configuration.",
	Long:  "Validate the metric simulation configuration file.",
	Args:  cobra.ExactArgs(1),
	Run:   check,
}

func init() {
	//.PersistentFlags().StringVar(&configFile, "config_file", configFile, "Configuration file to use")
	rootCmd.AddCommand(checkCmd)
}

func check(cmd *cobra.Command, args []string) {
	config, err := readAndValidateConfig(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config problem: %v", err)
		os.Exit(1)
	}

	configBytes, _ := yaml.Marshal(config)
	fmt.Printf(string(configBytes))
}
