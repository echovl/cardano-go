package cmd

import (
	"fmt"

	"github.com/echovl/cardano-wallet/db"
	"github.com/echovl/cardano-wallet/wallet"
	"github.com/spf13/cobra"
)

// listWalletCmd represents the listWallet command
var listWalletCmd = &cobra.Command{
	Use:     "list-wallets",
	Short:   "Print a list of known wallets",
	Aliases: []string{"lsw"},
	RunE: func(cmd *cobra.Command, args []string) error {
		bdb := db.NewBadgerDB()
		defer bdb.Close()

		wallets := wallet.GetWallets(bdb)

		fmt.Printf("%-18v %-9v %-9v\n", "ID", "NAME", "ADDRESS")
		for _, w := range wallets {
			addresses, err := w.Addresses(wallet.Mainnet)
			if err != nil {
				return err
			}
			fmt.Printf("%-18v %-9v %-9v\n", w.ID, w.Name, len(addresses))
			fmt.Println(w.ExternalChain.Childs[0].Xsk)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listWalletCmd)
}
