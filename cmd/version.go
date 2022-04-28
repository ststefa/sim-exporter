package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (

	// The proper version is automatically set to the contents of `version.txt` at link-time, see Makefile
	version = "dev"

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Long:  "Show the version of the exporter",
		Args:  cobra.NoArgs,
		Run:   doVersion,
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

// Any undesired but handled outcome is signaled by panicking with SimulationError
func doVersion(cmd *cobra.Command, args []string) {
	fmt.Println(version)
}
