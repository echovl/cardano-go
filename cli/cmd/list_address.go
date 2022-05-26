package cmd

import (
	"fmt"
	"strconv"

	"github.com/echovl/cardano-go/node"
	"github.com/echovl/cardano-go/types"
	"github.com/echovl/cardano-go/wallet"
	"github.com/spf13/cobra"
)

// listAddressCmd represents the listAddress command
var listAddressCmd = &cobra.Command{
	Use:     "list-address [wallet-id]",
	Short:   "Print a list of known wallet's addresses",
	Aliases: []string{"lsa"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		useTestnet, _ := cmd.Flags().GetBool("testnet")
		network := types.Mainnet
		if useTestnet {
			network = types.Testnet
		}

		opts := &wallet.Options{
			Node: node.NewCli(network),
		}
		client := wallet.NewClient(opts)
		defer client.Close()

		id := args[0]
		w, err := client.Wallet(id)
		if err != nil {
			return err
		}

		addresses := w.Addresses()
		fmt.Printf("%-25v %-9v\n", "PATH", "ADDRESS")
		for i, addr := range addresses {
			fmt.Printf("%-25v %-9v\n", "m/1852'/1815'/0'/0/"+strconv.Itoa(i), addr)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listAddressCmd)
	listAddressCmd.Flags().Bool("testnet", false, "Use testnet network")
}
