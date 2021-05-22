package cmd

import (
	"fmt"

	"github.com/echovl/cardano-go"
	"github.com/spf13/cobra"
)

var balanceCmd = &cobra.Command{
	Use:     "balance [cardano.id]",
	Short:   "Get cardano.s balance",
	Aliases: []string{"bal"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		useTestnet, err := cmd.Flags().GetBool("testnet")
		network := cardano.Mainnet
		if useTestnet {
			network = cardano.Testnet
		}

		id := cardano.WalletID(args[0])

		w, err := cardano.GetWallet(id, DefaultDb)
		if err != nil {
			return err
		}

		w.SetNetwork(network)
		w.SetNode(DefaultCardanoNode)

		balance, err := w.Balance()
		fmt.Printf("%-25v %-9v\n", "ASSET", "AMOUNT")
		fmt.Printf("%-25v %-9v\n", "Lovelace", balance)
		return err
	},
}

func init() {
	rootCmd.AddCommand(balanceCmd)
	balanceCmd.Flags().Bool("testnet", false, "Use testnet network")
}
