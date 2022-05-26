package cmd

import (
	"github.com/echovl/cardano-go/wallet"
	"github.com/spf13/cobra"
)

// deleteWalletCmd represents the deleteWallet command
var deleteWalletCmd = &cobra.Command{
	Use:     "delete-wallet [wallet-id]",
	Short:   "Delete a wallet",
	Aliases: []string{"delw"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := wallet.NewClient()
		id := args[0]
		return client.DeleteWallet(id)
	},
}

func init() {
	rootCmd.AddCommand(deleteWalletCmd)
}
