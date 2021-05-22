package cmd

import (
	"github.com/echovl/cardano-go"
	"github.com/spf13/cobra"
)

// TODO: Ask for password if present
// Experimental feature, only for testnet
var transferCmd = &cobra.Command{
	Use:   "transfer [wallet-id] [amount] [receiver-address]",
	Short: "Transfer an amount of lovelace to the given address",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		senderId := cardano.WalletID(args[0])
		receiver := cardano.Address(args[2])
		amountToTransfer, err := cardano.ParseUint64(args[1])
		if err != nil {
			return err
		}

		w, err := cardano.GetWallet(senderId, DefaultDb)
		if err != nil {
			return err
		}

		w.SetNetwork(cardano.Testnet)
		w.SetNode(DefaultCardanoNode)

		err = w.Transfer(receiver, amountToTransfer)

		return err
	},
}

func init() {
	rootCmd.AddCommand(transferCmd)
}
