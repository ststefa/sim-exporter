package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

func do(cmd *cobra.Command, args []string) {
	fmt.Println("version: " + version)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Long:  "Show the version of this exporter",
	Run:   do,
}
