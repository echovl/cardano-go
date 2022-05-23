package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tclairet/cardano-go"
)

// TODO: Ask for password if present
// Experimental feature, only for testnet
var transferCmd = &cobra.Command{
	Use:   "transfer [wallet-id] [amount] [receiver-address]",
	Short: "Transfer an amount of lovelace to the given address",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := cardano.NewClient()
		defer client.Close()
		senderId := args[0]
		receiver := cardano.Address(args[2])
		amount, err := cardano.ParseUint64(args[1])
		if err != nil {
			return err
		}
		w, err := client.Wallet(senderId)
		if err != nil {
			return err
		}
		w.SetNetwork(cardano.Testnet)
		err = w.Transfer(receiver, amount)
		return err
	},
}

func init() {
	rootCmd.AddCommand(transferCmd)
}
