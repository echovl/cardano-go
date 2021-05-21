package cmd

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"

	"github.com/echovl/cardano-wallet/db"
	"github.com/echovl/cardano-wallet/wallet"
	"github.com/fxamacker/cbor/v2"
	"github.com/spf13/cobra"
)

type TxJson struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	CborHex     string `json:"cborHex"`
}

var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "tx build example",
	RunE: func(cmd *cobra.Command, args []string) error {
		badger := db.NewBadgerDB()
		walletId := wallet.WalletID("wallet_8TPxCbfRyD")
		w, err := wallet.GetWallet(walletId, badger)
		if err != nil {
			return err
		}

		builder := wallet.NewTxBuilder(wallet.ProtocolParams{MinFeeA: 44, MinFeeB: 155381})

		xsk := w.ExternalChain.Childs[0].Xsk
		xvk := xsk.XVerificationKey()

		addresses, _ := w.Addresses(wallet.Testnet)
		txId, _ := hex.DecodeString("03c5301b0b739d867f49576d8a81690a354fdf917853044d53ccb432228e9c9d")
		utxoIndex := uint64(0)
		utxoAmount := uint64(1000 * 1e6)
		ttl := uint64(37224254)

		toAddress := wallet.Address("addr_test1vr228vewyrhp9ewkrnr7ju2heh6ja905mc23hv03f4r5a4ssjdkds")
		amountToSend := uint64(5 * 1e6)

		// Build transaction
		builder.AddInput(xvk, txId, utxoIndex, utxoAmount)
		builder.AddOutput(toAddress, amountToSend)
		builder.SetTtl(ttl)
		err = builder.AddFee(addresses[0])
		if err != nil {
			return err
		}
		builder.Sign(xsk)

		tx := builder.Build()
		txCborBytes, _ := cbor.Marshal(tx)

		// Tx json following cardano-cli format
		txJson := TxJson{
			Type:        "Tx MaryEra",
			Description: "",
			CborHex:     hex.EncodeToString(txCborBytes),
		}
		txJsonBytes, _ := json.Marshal(txJson)
		ioutil.WriteFile("tx_signed.json", txJsonBytes, 770)

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
