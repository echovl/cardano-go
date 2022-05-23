package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tclairet/cardano-go"
)

// deleteWalletCmd represents the deleteWallet command
var deleteWalletCmd = &cobra.Command{
	Use:     "delete-wallet [wallet-id]",
	Short:   "Delete a wallet",
	Aliases: []string{"delw"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := cardano.NewClient()
		id := args[0]
		return client.DeleteWallet(id)
	},
}

func init() {
	rootCmd.AddCommand(deleteWalletCmd)
}
