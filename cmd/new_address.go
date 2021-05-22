package cmd

import (
	"github.com/echovl/cardano-wallet/logger"
	"github.com/echovl/cardano-wallet/wallet"
	"github.com/spf13/cobra"
)

// newAddressCmd represents the address command
var newAddressCmd = &cobra.Command{
	Use:     "new-address [wallet-id]",
	Short:   "Create a new address",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"newa"},
	RunE: func(cmd *cobra.Command, args []string) error {
		useTestnet, err := cmd.Flags().GetBool("testnet")
		network := wallet.Mainnet
		if useTestnet {
			network = wallet.Testnet
		}

		walletID := wallet.WalletID(args[0])
		w, err := wallet.GetWallet(walletID, DefaultDb)
		if err != nil {
			return err
		}

		w.SetNetwork(network)

		addr := w.GenerateAddress()
		logger.Infow("New address created", "wallet", w.ID, "address", addr)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newAddressCmd)
	newAddressCmd.Flags().Bool("testnet", false, "Use testnet network")
}
