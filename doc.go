// Package cardano-go is a library for the Cardano Blockchain.
//
// Getting started
//
// The best way to get started working with cardano-go is to use `go get` to add the
// library.
//
//  go get github.com/echovl/cardano-go
//	go get github.com/echovl/cardano-go/node
//	go get github.com/echovl/cardano-go/tx
//
// Hello cardano-go
//
// This example shows how you can use cardano-go to query UTXOs.
//
//  package main
//
//  import (
//      "fmt"
//
//      "github.com/echovl/cardano-go/node/blockfrost"
//      "github.com/echovl/cardano-go/types"
//  )
//
//  func main() {
//      addr, err := types.NewAddress("addr1v9c7flddz5nqj448t78h5mdgau8p9ee24m7n62g2s48akkcjzfhw3")
//      if err != nil {
//          panic(err)
//      }
//
//      node := blockfrost.NewNode(types.Mainnet, "project-id")
//
//      utxos, err := node.UTXOs(addr)
//      if err != nil {
//          panic(err)
//      }
//
//      fmt.Println(utxos)
//  }
package cardano
