package cmd

import (
	"fmt"

	"github.com/echovl/cardano-go/node/cli"
	"github.com/echovl/cardano-go/types"
	"github.com/echovl/cardano-go/wallet"
	"github.com/spf13/cobra"
)

var balanceCmd = &cobra.Command{
	Use:     "balance [cardano.id]",
	Short:   "Get cardano.s balance",
	Aliases: []string{"bal"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		useTestnet, _ := cmd.Flags().GetBool("testnet")
		network := types.Mainnet
		if useTestnet {
			network = types.Testnet
		}

		opts := &wallet.Options{
			Node: cli.NewNode(network),
		}
		client := wallet.NewClient(opts)
		defer client.Close()

		id := args[0]
		w, err := client.Wallet(id)
		if err != nil {
			return err
		}

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
