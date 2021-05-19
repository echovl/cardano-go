package cmd

import (
	"github.com/echovl/cardano-wallet/db"
	"github.com/echovl/cardano-wallet/wallet"
	"github.com/spf13/cobra"
)

// deleteWalletCmd represents the deleteWallet command
var deleteWalletCmd = &cobra.Command{
	Use:     "delete-wallet [wallet-id]",
	Short:   "Delete a wallet",
	Aliases: []string{"delw"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bdb := db.NewBadgerDB()
		defer bdb.Close()

		walletId := wallet.WalletID(args[0])
		return wallet.DeleteWallet(walletId, bdb)
	},
}

func init() {
	rootCmd.AddCommand(deleteWalletCmd)
}
