package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"git.mgmt.innovo-cloud.de/operations-center/operationscenter-observability/sim-exporter/internal/metrics"
	"git.mgmt.innovo-cloud.de/operations-center/operationscenter-observability/sim-exporter/pkg/errors"

	"github.com/spf13/cobra"
)

var (
	outfile_help        = "Where to write the output"
	outfile      string = "/dev/stdout"

	maxdeviation_help     = "How many percent to deviate from converted value at most"
	maxdeviation      int = 10

	convertCmd = &cobra.Command{
		Use:   "convert <prometheus-scrape-file>",
		Short: "Parse prometheus-style scrape file and create simulator yaml config",
		Long:  "Parses data in prometheus scrape format read from <prometheus-scrape-file> and turns it into a yaml structure suitable as input for the simulator",
		Args:  cobra.ExactArgs(1),
		Run:   doConvert,
	}
)

func init() {
	convertCmd.Flags().StringVar(&outfile, "outfile", outfile, outfile_help)
	convertCmd.Flags().IntVar(&maxdeviation, "maxdeviation", maxdeviation, maxdeviation_help)

	rootCmd.AddCommand(convertCmd)
}

// Any undesired but handled outcome is signaled by panicking with SimulationError
func doConvert(cmd *cobra.Command, args []string) {
	//var config Configuration

	yamlData, err := metrics.ConvertScrapefileToYaml(args[0], maxdeviation)
	if err != nil {
		panic(&errors.SimulationError{Err: err.Error()})
	}

	err = ioutil.WriteFile(outfile, yamlData, 0644)
	if err != nil {
		panic(&errors.SimulationError{Err: err.Error()})
	} else {
		fmt.Fprintf(os.Stderr, "Wrote config to %v\n", outfile)
	}
}
