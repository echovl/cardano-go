package cmd

import (
	"github.com/echovl/cardano-go/node/blockfrost"
	"github.com/echovl/cardano-go/types"
	"github.com/echovl/cardano-go/wallet"
	"github.com/spf13/cobra"
)

// newAddressCmd represents the address command
var newAddressCmd = &cobra.Command{
	Use:     "new-address [wallet-id]",
	Short:   "Create a new address",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"newa"},
	RunE: func(cmd *cobra.Command, args []string) error {
		useTestnet, _ := cmd.Flags().GetBool("testnet")
		network := types.Mainnet
		if useTestnet {
			network = types.Testnet
		}

		node := blockfrost.NewNode(network, cfg.BlockfrostProjectID)
		opts := &wallet.Options{Node: node}
		client := wallet.NewClient(opts)
		defer client.Close()

		id := args[0]
		w, err := client.Wallet(id)
		if err != nil {
			return err
		}
		w.AddAddress()
		client.SaveWallet(w)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newAddressCmd)
	newAddressCmd.Flags().Bool("testnet", false, "Use testnet network")
}
