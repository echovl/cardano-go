package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tclairet/cardano-go"
)

var newWalletCmd = &cobra.Command{
	Use:   "new-wallet [wallet-name]",
	Short: "Create or restore a wallet",
	Long: `Create or restore a wallet. If the mnemonic flag is present 
it will restore a wallet using the mnemonic and password.`,
	Aliases: []string{"neww"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := cardano.NewClient()
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
}
