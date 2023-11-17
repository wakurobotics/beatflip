package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// the app's version. This will be set on build.
var Version string = "dev"

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
