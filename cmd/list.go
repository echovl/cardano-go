package cmd

import (
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List wallets and addresses",
}

func init() {
	rootCmd.AddCommand(listCmd)
}
