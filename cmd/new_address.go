package cmd

import (
	"fmt"

	"github.com/echovl/cardano-wallet/db"
	"github.com/echovl/cardano-wallet/wallet"
	"github.com/spf13/cobra"
)

// addressCmd represents the address command
var addressCmd = &cobra.Command{
	Use:   "address [wallet-id]",
	Short: "Create a new address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		useTestnet, err := cmd.Flags().GetBool("testnet")
		network := wallet.Mainnet
		if useTestnet {
			network = wallet.Testnet
		}

		walletID := wallet.WalletID(args[0])
		bdb := db.NewBadgerDB()
		defer bdb.Close()

		w, err := wallet.GetWallet(walletID, bdb)
		if err != nil {
			return err
		}

		addr, err := w.GenerateAddress(network)
		if err != nil {
			return err
		}

		fmt.Println("New address", addr)

		return nil
	},
}

func init() {
	newCmd.AddCommand(addressCmd)
	addressCmd.Flags().Bool("testnet", false, "Use testnet network")
}
