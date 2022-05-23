package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tclairet/cardano-go"
)

// newAddressCmd represents the address command
var newAddressCmd = &cobra.Command{
	Use:     "new-address [wallet-id]",
	Short:   "Create a new address",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"newa"},
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
		w.AddAddress()
		client.SaveWallet(w)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newAddressCmd)
	newAddressCmd.Flags().Bool("testnet", false, "Use testnet network")
}
