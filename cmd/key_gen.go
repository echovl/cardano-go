package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/echovl/cardano-wallet/crypto"
	"github.com/spf13/cobra"
	"github.com/tyler-smith/go-bip39"
)

// keyGenCmd represents the keyGen command
var keyGenCmd = &cobra.Command{
	Use:   "key-gen",
	Short: "Generate a new key pair",
	Run: func(cmd *cobra.Command, args []string) {
		entropy, err := bip39.NewEntropy(20 * 8)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		mnemonic, err := crypto.GenerateMnemonic(entropy)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		xpriv := crypto.GenerateMasterKey(entropy, "test")

		fmt.Println("entropy: ", hex.EncodeToString(entropy))
		fmt.Println("mnemonic: ", mnemonic)
		fmt.Println("xpriv: ", hex.EncodeToString(xpriv))

		xpriv1, _ := crypto.DerivePrivateKey(xpriv, 0x81000000)
		fmt.Println("xpriv1: ", hex.EncodeToString(xpriv1))
	},
}

func init() {
	rootCmd.AddCommand(keyGenCmd)
}
