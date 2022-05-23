package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tclairet/cardano-go"
)

// listAddressCmd represents the listAddress command
var listAddressCmd = &cobra.Command{
	Use:     "list-address [wallet-id]",
	Short:   "Print a list of known wallet's addresses",
	Aliases: []string{"lsa"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := cardano.NewClient()
		defer client.Close()

		useTestnet, err := cmd.Flags().GetBool("testnet")
		network := cardano.Mainnet
		if useTestnet {
			network = cardano.Testnet
		}

		id := args[0]
		w, err := client.Wallet(id)
		if err != nil {
			return err
		}
		w.SetNetwork(network)

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
