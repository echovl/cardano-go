package cmd

import (
	"github.com/echovl/cardano-wallet/db"
	"github.com/echovl/cardano-wallet/provider"
	"github.com/echovl/cardano-wallet/wallet"
	"github.com/spf13/cobra"
)

// TODO: Ask for password if present
// Experimental feature, only for testnet
var transferCmd = &cobra.Command{
	Use:   "transfer [wallet-id] [amount] [receiver-address]",
	Short: "Transfer an amount of lovelace to the given address",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		senderId := wallet.WalletID(args[0])
		receiver := wallet.Address(args[2])
		amountToTransfer, err := wallet.ParseUint64(args[1])
		if err != nil {
			return err
		}

		badger := db.NewBadgerDB()
		provider := &provider.NodeCli{}

		w, err := wallet.GetWallet(senderId, badger)
		if err != nil {
			return err
		}

		w.SetNetwork(wallet.Testnet)
		w.SetProvider(provider)

		err = w.Transfer(receiver, amountToTransfer)

		return err
	},
}

func init() {
	rootCmd.AddCommand(transferCmd)
}
