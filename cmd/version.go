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
		Long:  "Show the version of this exporter",
		Args:  cobra.NoArgs,
		Run:   do,
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

func do(cmd *cobra.Command, args []string) {
	fmt.Println(version)
}
