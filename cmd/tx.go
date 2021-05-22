package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/echovl/cardano-wallet/provider"
	"github.com/echovl/cardano-wallet/wallet"
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
		toAddress := wallet.Address("addr_test1vr228vewyrhp9ewkrnr7ju2heh6ja905mc23hv03f4r5a4ssjdkds")

		provider := provider.NodeCli{}
		utxos, err := provider.QueryUtxos(toAddress)
		tip, err := provider.QueryTip()

		fmt.Println(utxos, tip)

		return err
	},
}

func pretty(v interface{}) string {
	bytes, _ := json.MarshalIndent(v, "", "  ")
	return string(bytes)
}

func init() {
	rootCmd.AddCommand(txCmd)
}
