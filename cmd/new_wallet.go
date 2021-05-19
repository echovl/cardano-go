package cmd

import (
	"fmt"
	"strings"

	"github.com/echovl/cardano-wallet/db"
	"github.com/echovl/cardano-wallet/wallet"
	"github.com/spf13/cobra"
)

// newWalletCmd represents the wallet command
var newWalletCmd = &cobra.Command{
	Use:   "new-wallet [wallet-name]",
	Short: "Create or restore a wallet",
	Long: `Create or restore a wallet. If the mnemonic flag is present 
it will restore a wallet using the mnemonic and password.`,
	Aliases: []string{"neww"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		password, _ := cmd.Flags().GetString("password")
		mnemonic, _ := cmd.Flags().GetStringSlice("mnemonic")
		bdb := db.NewBadgerDB()
		defer bdb.Close()

		if len(mnemonic) == 0 {
			_, mnemonic, _ := wallet.AddWallet(args[0], password, bdb)

			fmt.Printf("mnemonic: %v\n", mnemonic)
		} else {
			wallet.RestoreWallet(args[0], strings.Join(mnemonic, " "), password, bdb)
		}
	},
}

func init() {
	rootCmd.AddCommand(newWalletCmd)

	newWalletCmd.Flags().StringP("password", "p", "", "A list of mnemonic words")
	newWalletCmd.Flags().StringSliceP("mnemonic", "m", nil, "Password to lock and protect the wallet")
}
