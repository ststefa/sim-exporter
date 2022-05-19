package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"git.mgmt.innovo-cloud.de/obs/sim-exporter/pkg/errors"
	"git.mgmt.innovo-cloud.de/obs/sim-exporter/pkg/metrics"

	"github.com/spf13/cobra"
)

var (
	outfile_help = "Where to write the output to"
	outfile      = "/dev/stdout"

	maxdeviation_help = "How many percent to deviate from converted value at most (symmetrically in positive and negative direction)"
	maxdeviation      = 50

	function_help = "Function by which to modify value over time. Comma separated string consisting of one or more of rand, asc, desc, sin."
	function      = "rand,asc,desc,sin"

	interval_help = "Minimum-maximum duration of function interval"
	interval      = "10m-2h"

	honorpct_help = "Use absolute deviation for metrics containing this string (comma separated list of substrings)"
	honorpct      = "percent"

	convertCmd = &cobra.Command{
		Use:     "convert <prometheus-scrape-file>",
		Short:   "Parse prometheus-style scrape file and create simulator yaml config",
		Long:    "Parses data in prometheus scrape format read from <prometheus-scrape-file> and turns it into a yaml structure suitable as input for the simulator",
		Args:    cobra.ExactArgs(1),
		PreRunE: validateConvert,
		Run:     doConvert,
	}
)

func init() {
	convertCmd.Flags().StringVarP(&outfile, "outfile", "o", outfile, outfile_help)
	convertCmd.Flags().IntVarP(&maxdeviation, "maxdeviation", "d", maxdeviation, maxdeviation_help)
	convertCmd.Flags().StringVarP(&function, "function", "f", function, function_help)
	convertCmd.Flags().StringVarP(&interval, "interval", "i", interval, interval_help)
	convertCmd.Flags().StringVarP(&honorpct, "honorpct", "p", honorpct, honorpct_help)

	rootCmd.AddCommand(convertCmd)
}

func validateConvert(cmd *cobra.Command, args []string) error {

	// Validate deviation
	if port < 0 || maxdeviation > 100 {
		return fmt.Errorf("maxdeviation must be in range 0-100")
	}

	// More complex validations performed in ScrapefileToCollection
	return nil
}

// Any undesired but handled outcome is signaled by panicking with SimulationError
func doConvert(cmd *cobra.Command, args []string) {

	collection, err := metrics.ScrapefileToCollection(args[0], maxdeviation, function, interval, honorpct)
	if err != nil {
		panic(&errors.SimulationError{Err: err.Error()})
	}

	yamlData, err := yaml.Marshal(collection)
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
