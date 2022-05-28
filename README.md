# cardano-go

cardano-go is a library for creating go applicactions that interact with the Cardano Blockchain. [WIP]

## Installation

```
$ go get github.com/echovl/cardano-node
$ go get github.com/echovl/cardano-node/node
$ go get github.com/echovl/cardano-node/tx
```

## Quickstart

```go
import (
	"fmt"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/node/blockfrost"
	"github.com/echovl/cardano-go/tx"
	"github.com/echovl/cardano-go/types"
)

func main() {
	// Node using blockfrost
	node := blockfrost.NewNode(types.Mainnet, "project-id")

	// New transaction builder
	builder := tx.NewTxBuilder(types.ProtocolParams{
		MinimumUtxoValue: 1000000,
		MinFeeA:          44,
		MinFeeB:          155381,
	})

	// Fetch node tip information
	tip, err := node.Tip()
	if err != nil {
		panic(err)
	}

	xprv, err := crypto.NewXPrvFromBech32("xpriv")
	if err != nil {
		panic(err)
	}
	txInputHash, err := types.NewHash32("tx-hash")
	if err != nil {
		panic(err)
	}
	sender, err := types.NewAddress("addr")
	if err != nil {
		panic(err)
	}
	receiver, err := types.NewAddress("addr")
	if err != nil {
		panic(err)
	}

	// Build inputs and ouputs
	txInput := tx.TransactionInput{TxHash: txInputHash, Index: 0, Amount: types.Coin(14838997)}
	txOutput := tx.TransactionOutput{Address: receiver, Amount: 1000000}

	// Add inputs and outputs
	builder.AddInputs(txInput)
	builder.AddOutputs(txOutput)

	// Add fee and change output
	builder.AddFee(sender)

	// Set time to live
	builder.SetTTL(tip.Slot + 100)

	// Sign transaction
	builder.Sign(xprv)

	// Build transaction
	tx, err := builder.Build()
	if err != nil {
		panic(err)
	}

	// Submit transaction to node
	txHash, err := node.SubmitTx(tx)
	if err != nil {
		panic(err)
	}

	fmt.Println(txHash)
}
```

## cwallet

This package provides a cli `cwallet` for wallet management. It supports:

- hierarchical deterministic wallets
- fetching balances
- transfering ADA

### Configuration

`cwallet` uses blockfrost as a Node provider. It requires a config file in `$XDG_CONFIG_HOME/cwallet.yml`.
Example:

```yaml
blockfrost_project_id: "project-id"
```

### Installation

```
$ git clone github.com/echovl/cardano-go
$ make && sudo make install
```

### Usage

Wallet creation:

```
$ cwallet new-wallet jhon
mnemonic: various find knee churn bicycle current midnight visit artist help soon flower venture wasp problem
```

List wallet address:

```
$ cwallet list-address jhon
PATH                      ADDRESS
m/1852'/1815'/0'/0/0      addr1vxzfs9dj365gcdmv6dwj7auewf624ghwrtduecu37hrxsyst8gvu2
```

Send ADA:

```
$ cwallet transfer echo addr1vxzfs9dj365gcdmv6dwj7auewf624ghwrtduecu37hrxsyst8gvu2 2000000
fd3a7d6e9742fd9ddba2bd1740fa994f5c93a4f59bf88dc5f81d8d7413c5b3a9
```
