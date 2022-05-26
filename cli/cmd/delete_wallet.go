package cmd

import (
	"github.com/echovl/cardano-go/node"
	"github.com/echovl/cardano-go/types"
	"github.com/echovl/cardano-go/wallet"
	"github.com/spf13/cobra"
)

// deleteWalletCmd represents the deleteWallet command
var deleteWalletCmd = &cobra.Command{
	Use:     "delete-wallet [wallet-id]",
	Short:   "Delete a wallet",
	Aliases: []string{"delw"},
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
		id := args[0]
		return client.DeleteWallet(id)
	},
}

func init() {
	rootCmd.AddCommand(deleteWalletCmd)
	deleteWalletCmd.Flags().Bool("testnet", false, "Use testnet network")
}
