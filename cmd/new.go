package cmd

import (
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create wallets and addresses",
}

func init() {
	rootCmd.AddCommand(newCmd)
}
