package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helloCmd = &cobra.Command{
	Use:     "hello",
	Short:   "Print Hello World",
	GroupID: "system",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "Hello World")
	},
}

func init() {
	rootCmd.AddCommand(helloCmd)
}
