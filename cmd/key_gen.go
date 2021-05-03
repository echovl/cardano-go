package cmd

import (
	"github.com/spf13/cobra"
)

// keyGenCmd represents the keyGen command
var keyGenCmd = &cobra.Command{
	Use:   "key-gen",
	Short: "Generate a new key pair",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(keyGenCmd)
}
