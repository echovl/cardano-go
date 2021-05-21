package cmd

import (
	"encoding/json"
	"log"

	"github.com/echovl/cardano-wallet/crypto"
	"github.com/echovl/cardano-wallet/wallet"
	"github.com/spf13/cobra"
)

// txCmd represents the tx command
var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "tx build demo",
	RunE: func(cmd *cobra.Command, args []string) error {
		builder := wallet.NewTxBuilder(wallet.ProtocolParams{MinFeeA: 44, MinFeeB: 155381})

		xsk := crypto.XSigningKey([]byte{
			0, 237, 225, 96, 167, 183, 203, 122, 22, 221, 248, 249, 240, 177, 80, 198, 37, 204, 119, 73, 45, 226, 57, 58, 224, 185, 85, 164, 227, 173, 70, 79, 64, 243, 101, 168, 111, 20, 206, 86, 113, 111, 154, 66, 61, 66, 101, 117, 174, 0, 15, 171, 123, 21, 98, 174, 13, 161, 244, 47, 97, 95, 134, 20, 123, 37, 73, 112, 113, 60, 207, 51, 101, 180, 225, 3, 55, 95, 172, 17, 167, 252, 67, 207, 55, 165, 12, 169, 149, 158, 67, 195, 223, 250, 251, 230,
		})
		xvk := xsk.XVerificationKey()
		txId := []byte{
			23, 43, 54, 23,
			23, 43, 54, 23,
			23, 43, 54, 23,
			23, 43, 54, 23,
			23, 43, 54, 23,
			23, 43, 54, 23,
			23, 43, 54, 23,
			23, 43, 54, 23,
		}

		builder.AddInput(xvk, txId, 0, 10*1e6)
		builder.AddOutput("addr1v8fykcd00h5l49qyeq6r0s86mvyl6wug9jwuz8dpyv69tpc9wyn2g", 5*1e6)
		builder.SetTtl(100)
		err := builder.AddFee("addr1v8fykcd00h5l49qyeq6r0s86mvyl6wug9jwuz8dpyv69tpc9wyn2g")
		if err != nil {
			return err
		}
		builder.Sign(xsk)

		tx := builder.Build()

		log.Println("body: ", pretty(tx.Body))
		log.Println("witness: ", pretty(tx.WitnessSet))

		txValid := xvk.Verify(tx.Body.Bytes(), tx.WitnessSet.VKeyWitnessSet[0].Signature)

		log.Println("transaction is valid?", txValid)

		return nil
	},
}

func pretty(v interface{}) string {
	bytes, _ := json.MarshalIndent(v, "", "  ")
	return string(bytes)
}

func init() {
	rootCmd.AddCommand(txCmd)
}
