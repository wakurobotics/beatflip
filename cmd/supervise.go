package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wakurobotics/beatflip/supervisor"
)

var superviseCmd = &cobra.Command{
	Use: "supervise",
	RunE: func(cmd *cobra.Command, args []string) error {
		return supervisor.Supervise()
	},
}

func init() {
	rootCmd.AddCommand(superviseCmd)
}
