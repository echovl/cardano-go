package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stakecore/cardano-wallet-go/internal/bech32"
)

// helloCmd represents the hello command
var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello called")

		rootKey := "root_xsk1spewah0xnqc5lk4hxcpunlmyv76naz69kwm9nvpjzal8fepwye8mu222f0pq39wam0mqf3wgk28xvjl3rn0fheuql272wp2qlgu88tmzjmjfckn90lf52l8cfysy66k53dt2dzqjusmzmkk7tfltq4grku60xxg2"
		hrp, decoded, err := bech32.DecodeToBase256(rootKey)
		if err != nil {
			fmt.Println("Error:", err)
		}

		fmt.Println("Hrp:", hrp)
		fmt.Println("Data:", hex.EncodeToString(decoded))
		fmt.Println("Data length:", len(decoded))

		rootKey2, _ := bech32.EncodeFromBase256("root_xsk", decoded)
		fmt.Println("Encoded again:", rootKey2)

	},
}

func init() {
	rootCmd.AddCommand(helloCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// helloCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// helloCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
