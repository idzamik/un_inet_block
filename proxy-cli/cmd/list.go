package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured servers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("list")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
