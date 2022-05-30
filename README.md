# cardano-go

cardano-go is a library for creating go applicactions that interact with the Cardano Blockchain. [WIP]

## Installation

```
$ go get github.com/echovl/cardano-node
$ go get github.com/echovl/cardano-node/node
$ go get github.com/echovl/cardano-node/tx
```

## Usage

### Get protocol parameters

```go
package main

import (
	"fmt"

	"github.com/echovl/cardano-go/node/blockfrost"
	"github.com/echovl/cardano-go/types"
)

func main() {
	node := blockfrost.NewNode(types.Mainnet, "project-id")

	pparams, err := node.ProtocolParams()
	if err != nil {
		panic(err)
	}

	fmt.Println(pparams)
}
```

### Create simple transaction

```go
package main

import (
	"fmt"

	"github.com/echovl/cardano-go/tx"
	"github.com/echovl/cardano-go/types"
)

func main() {
	txBuilder := tx.NewTxBuilder(&tx.ProtocolParams{})

	txInput, err := tx.NewTxInput("txhash", 0, types.Coin(2000000))
	if err != nil {
		panic(err)
	}
	txOut, err := tx.NewTxOutput("addr", types.Coin(1300000))
	if err != nil {
		panic(err)
	}

	txBuilder.AddInputs(txInput)
	txBuilder.AddOutputs(txOut)
	txBuilder.SetTTL(100000)
	txBuilder.SetFee(types.Coin(160000))
	txBuilder.Sign("xprv")

	tx, err := txBuilder.Build()
	if err != nil {
		panic(err)
	}

	fmt.Println(tx.Hex())
}
```

### Set fee and change output

```go
package main

import (
	"fmt"

	"github.com/echovl/cardano-go/tx"
)

func main() {
	txBuilder := tx.NewTxBuilder(&tx.ProtocolParams{})

	// Transaction should be signed at this point
	err := txBuilder.AddChangeIfNeeded("addr")
	if err != nil {
		panic(err)
	}
}
```

### Submit transaction

```go
package main

import (
	"fmt"

	"github.com/echovl/cardano-go/node/blockfrost"
	"github.com/echovl/cardano-go/tx"
	"github.com/echovl/cardano-go/types"
)

func main() {
	node := blockfrost.NewNode(types.Mainnet, "project-id")

	txHash, err := node.SubmitTx(&tx.Tx{})
	if err != nil {
		panic(err)
	}

	fmt.Println(txHash)
}
```

### Add transaction metadata

```go
package main

import (
	"github.com/echovl/cardano-go/tx"
)

func main() {
	txBuilder := tx.NewTxBuilder(&tx.ProtocolParams{})

	txBuilder.AddAuxiliaryData(&tx.AuxiliaryData{
		Metadata: tx.Metadata{
			0: map[string]interface{}{
				"hello": "cardano-go",
			},
		},
	})
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
