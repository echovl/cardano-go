package cmd

import (
	"fmt"
	"strings"

	"github.com/echovl/cardano-go/node/blockfrost"
	"github.com/echovl/cardano-go/types"
	"github.com/echovl/cardano-go/wallet"
	"github.com/spf13/cobra"
)

var newWalletCmd = &cobra.Command{
	Use:   "new-wallet [name]",
	Short: "Create or restore a wallet",
	Long: `Create or restore a wallet. If the mnemonic flag is present 
it will restore a wallet using the mnemonic and password.`,
	Aliases: []string{"neww"},
	Args:    cobra.ExactArgs(1),
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
		password, _ := cmd.Flags().GetString("password")
		mnemonic, _ := cmd.Flags().GetStringSlice("mnemonic")
		name := args[0]

		if len(mnemonic) == 0 {
			_, mnemonic, err := client.CreateWallet(name, password)
			if err != nil {
				return err
			}
			fmt.Printf("mnemonic: %v\n", mnemonic)
		} else {
			_, err := client.RestoreWallet(name, password, strings.Join(mnemonic, " "))
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newWalletCmd)

	newWalletCmd.Flags().StringP("password", "p", "", "A list of mnemonic words")
	newWalletCmd.Flags().StringSliceP("mnemonic", "m", nil, "Password to lock and protect the wallet")
	newWalletCmd.Flags().Bool("testnet", false, "Use testnet network")
}
