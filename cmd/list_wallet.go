package cmd

import (
	"fmt"

	"github.com/echovl/cardano-wallet/db"
	"github.com/echovl/cardano-wallet/wallet"
	"github.com/spf13/cobra"
)

// listWalletCmd represents the listWallet command
var listWalletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Print a list of known wallets",
	RunE: func(cmd *cobra.Command, args []string) error {
		bdb := db.NewBadgerDB()
		defer bdb.Close()

		wallets := wallet.GetWallets(bdb)

		fmt.Printf("%-18v %-9v %-9v\n", "ID", "NAME", "ADDRESS")
		for _, v := range wallets {
			addresses, err := v.Addresses(wallet.Mainnet)
			if err != nil {
				return err
			}
			fmt.Printf("%-18v %-9v %-9v\n", v.ID, v.Name, len(addresses))
		}
		return nil
	},
}

func init() {
	listCmd.AddCommand(listWalletCmd)
}
