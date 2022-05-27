package cmd

import (
	"fmt"

	"github.com/echovl/cardano-go/node/cli"
	"github.com/echovl/cardano-go/types"
	"github.com/echovl/cardano-go/wallet"
	"github.com/spf13/cobra"
)

// listWalletCmd represents the listWallet command
var listWalletCmd = &cobra.Command{
	Use:     "list-wallets",
	Short:   "Print a list of known wallets",
	Aliases: []string{"lsw"},
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
		wallets, err := client.Wallets()
		if err != nil {
			return err
		}
		fmt.Printf("%-18v %-9v %-9v\n", "ID", "NAME", "ADDRESS")
		for _, w := range wallets {
			addresses := w.Addresses()
			fmt.Printf("%-18v %-9v %-9v\n", w.ID, w.Name, len(addresses))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listWalletCmd)
	listWalletCmd.Flags().Bool("testnet", false, "Use testnet network")
}
