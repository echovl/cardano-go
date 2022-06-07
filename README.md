# cardano-go

[![PkgGoDev](https://pkg.go.dev/badge/github.com/echovl/cardano-go)](https://pkg.go.dev/github.com/echovl/cardano-go) ![ci](https://github.com/echovl/cardano-go/workflows/test/badge.svg)

cardano-go is a library for creating go applicactions that interact with the Cardano Blockchain. [WIP]

## Installation

```
$ go get github.com/echovl/cardano-go
```

## Usage

### Get protocol parameters

```go
package main

import (
	"fmt"

	"github.com/echovl/cardano-go"
	"github.com/echovl/cardano-go/blockfrost"
)

func main() {
	node := blockfrost.NewNode(cardano.Mainnet, "project-id")

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

	"github.com/echovl/cardano-go"
	"github.com/echovl/cardano-go/crypto"
)

func main() {
	txBuilder := cardano.NewTxBuilder(&cardano.ProtocolParams{})

	sender, err := cardano.NewAddress("addr")
	if err != nil {
		panic(err)
	}
	receiver, err := cardano.NewAddress("addr")
	if err != nil {
		panic(err)
	}
	sk, err := crypto.NewPrvKey("addr_sk")
	if err != nil {
		panic(err)
	}
	txHash, err := cardano.NewHash32("txhash")
	if err != nil {
		panic(err)
	}

	txInput := cardano.NewTxInput(txHash, 0, cardano.NewValue(20e6))
	txOut := cardano.NewTxOutput(receiver, cardano.NewValue(10e6))

	txBuilder.AddInputs(txInput)
	txBuilder.AddOutputs(txOut)
	txBuilder.SetTTL(100000)
	txBuilder.AddChangeIfNeeded(sender)
	txBuilder.Sign(sk)

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
	"github.com/echovl/cardano-go"
)

func main() {
	txBuilder := cardano.NewTxBuilder(&cardano.ProtocolParams{})

	changeAddr, err := cardano.NewAddress("addr")
	if err != nil {
		panic(err)
	}

	txBuilder.AddChangeIfNeeded(changeAddr)
}
```

### Submit transaction

```go
package main

import (
	"fmt"

	"github.com/echovl/cardano-go"
	"github.com/echovl/cardano-go/blockfrost"
)

func main() {
	node := blockfrost.NewNode(cardano.Mainnet, "project-id")

	txHash, err := node.SubmitTx(&cardano.Tx{})
	if err != nil {
		panic(err)
	}

	fmt.Println(txHash)
}
```

### Add transaction metadata

```go
package main

import "github.com/echovl/cardano-go"

func main() {
	txBuilder := cardano.NewTxBuilder(&cardano.ProtocolParams{})

	txBuilder.AddAuxiliaryData(&cardano.AuxiliaryData{
		Metadata: cardano.Metadata{
			0: map[string]interface{}{
				"hello": "cardano-go",
			},
		},
	})
}
```

### Stake key registration

```go
package main

import (
	"github.com/echovl/cardano-go"
	"github.com/echovl/cardano-go/crypto"
)

func main() {
	txBuilder := cardano.NewTxBuilder(&cardano.ProtocolParams{})

	stakeKey, err := crypto.NewPrvKey("stake_sk")
	if err != nil {
		panic(err)
	}

	stakeRegCert, err := cardano.NewStakeRegistrationCertificate(stakeKey.PubKey())
	if err != nil {
		panic(err)
	}

	txBuilder.AddCertificate(stakeRegCert)
}
```

### Stake delegation

```go
package main

import (
	"github.com/echovl/cardano-go"
	"github.com/echovl/cardano-go/crypto"
)

func main() {
	txBuilder := cardano.NewTxBuilder(&cardano.ProtocolParams{})

	stakeKey, err := crypto.NewPrvKey("stake_sk")
	if err != nil {
		panic(err)
	}

	poolKeyHash, err := cardano.NewHash28("2d6765748cc86efe862f5abeb0c0271f91d368d300123ecedc078ef2")
    if err != nil {
        panic(err)
    }
	stakeDelCert, err := cardano.NewStakeDelegationCertificate(stakeKey.PubKey(), poolKeyHash)
	if err != nil {
		panic(err)
	}

	txBuilder.AddCertificate(stakeDelCert)
}
```

### Mint native tokens

```go
package main

import (
	"math/big"

	"github.com/echovl/cardano-go"
	"github.com/echovl/cardano-go/crypto"
)

func main() {
	txBuilder := cardano.NewTxBuilder(&cardano.ProtocolParams{})

	policyKey, err := crypto.NewPrvKey("policy_sk")
	if err != nil {
		panic(err)
	}
	policyScript, err := cardano.NewScriptPubKey(policyKey.PubKey())
	if err != nil {
		panic(err)
	}
	policyID, err := cardano.NewPolicyID(policyScript)
	if err != nil {
		panic(err)
	}

	newAsset := cardano.NewMint().
		Set(
			policyID,
			cardano.NewMintAssets().
				Set(cardano.NewAssetName("cardanogo"), big.NewInt(1e9)),
		)

	addr, err := cardano.NewAddress("addr")
	if err != nil {
		panic(err)
	}

	txHashIn, err := cardano.NewHash32("txhash")
	if err != nil {
		panic(err)
	}

	txBuilder.AddInputs(
		cardano.NewTxInput(txHashIn, 0, cardano.NewValue(10e6)),
	)
	txBuilder.AddOutputs(
		cardano.NewTxOutput(addr, cardano.NewValueWithAssets(10e6, newAsset.MultiAsset())),
	)

	txBuilder.AddNativeScript(policyScript)
	txBuilder.Mint(newAsset)
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
