package cmd

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

//go:embed .beatflip.yml
var configTemplate []byte

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize a new configuration file in the current directoey",
	Run: func(cmd *cobra.Command, args []string) {
		outfile := ".beatflip.yml"
		if cfgFile != "" {
			outfile = cfgFile
		}

		_, err := os.Stat(outfile)
		if err == nil || !errors.Is(err, os.ErrNotExist) {
			fmt.Printf("The file '%s' already exists. Overwrite it? (y/N): ", outfile)
			reader := bufio.NewReader(os.Stdin)
			text, err := reader.ReadString('\n')
			cobra.CheckErr(err)
			if text[0] != 'y' && text[0] != 'Y' {
				os.Exit(0)
			}
		}

		err = os.WriteFile(outfile, configTemplate, 0644)
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
