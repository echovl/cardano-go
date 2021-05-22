package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("balance called")
	},
}

func init() {
	rootCmd.AddCommand(balanceCmd)
}
