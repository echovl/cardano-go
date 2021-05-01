package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stakecore/cardano-wallet-go/crypto"
)

// keyGenCmd represents the keyGen command
var keyGenCmd = &cobra.Command{
	Use:   "key-gen",
	Short: "Generate a new key pair",
	Run: func(cmd *cobra.Command, args []string) {
		kp, _ := crypto.NewKeyPair()

		fmt.Printf("New key pair: \npriv: %v \npub: %v\n", hex.EncodeToString(kp.Priv), hex.EncodeToString(kp.Pub))
		fmt.Println("key-gen called")

		kp.Derive(0x82000000)
	},
}

func init() {
	rootCmd.AddCommand(keyGenCmd)
}
