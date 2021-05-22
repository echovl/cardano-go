package cmd

import (
	"fmt"

	"github.com/echovl/cardano-wallet/db"
	"github.com/echovl/cardano-wallet/provider"
	"github.com/echovl/cardano-wallet/wallet"
	"github.com/spf13/cobra"
)

var balanceCmd = &cobra.Command{
	Use:     "balance [wallet-id]",
	Short:   "Get wallet's balance",
	Aliases: []string{"bal"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		useTestnet, err := cmd.Flags().GetBool("testnet")
		network := wallet.Mainnet
		if useTestnet {
			network = wallet.Testnet
		}

		id := wallet.WalletID(args[0])
		badger := db.NewBadgerDB()
		provider := &provider.NodeCli{}
		w, err := wallet.GetWallet(id, badger)
		if err != nil {
			return err
		}

		balance, err := w.Balance(provider, network)
		fmt.Printf("%-25v %-9v\n", "ASSET", "AMOUNT")
		fmt.Printf("%-25v %-9v\n", "Lovelace", balance)
		return err
	},
}

func init() {
	rootCmd.AddCommand(balanceCmd)
	balanceCmd.Flags().Bool("testnet", false, "Use testnet network")
}
