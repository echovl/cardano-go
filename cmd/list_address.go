/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"strconv"

	"github.com/echovl/cardano-wallet/db"
	"github.com/echovl/cardano-wallet/wallet"
	"github.com/spf13/cobra"
)

// listAddressCmd represents the listAddress command
var listAddressCmd = &cobra.Command{
	Use:   "address [wallet-id]",
	Short: "Print a list of known wallet's addresses",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		walletID := wallet.WalletID(args[0])
		bdb := db.NewBadgerDB()
		defer bdb.Close()

		w, err := wallet.GetWallet(walletID, bdb)
		if err != nil {
			return err
		}
		addresses, err := w.Addresses()
		fmt.Printf("%-25v %-9v\n", "PATH", "ADDRESS")
		for i, addr := range addresses {
			fmt.Printf("%-25v %-9v\n", "m/1852'/1815'/0'/0/"+strconv.Itoa(i), addr)
		}

		return nil
	},
}

func init() {
	listCmd.AddCommand(listAddressCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listAddressCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listAddressCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
