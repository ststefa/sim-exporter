package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check <file.yaml>",
	Short: "Validate simulation config in <file.yaml>",
	Long:  "Validate the metric simulation configuration <file.yaml>",
	Args:  cobra.ExactArgs(1),
	Run:   doCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

// Any undesired but handled outcome is signaled by panicking with SimulationError
func doCheck(cmd *cobra.Command, args []string) {
	config, err := loadAndValidateConfiguration(args[0])
	if err != nil {
		panic(&SimulationError{err.Error()})
	}

	err = setupMetricsCollection(config)
	if err != nil {
		panic(&SimulationError{err.Error()})
	}

	fmt.Printf("%v validated successfully\n", args[0])
}
