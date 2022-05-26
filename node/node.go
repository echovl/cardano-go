package node

import (
	"github.com/echovl/cardano-go/tx"
	"github.com/echovl/cardano-go/types"
)

const (
	testnetMagic = 1097911063
)

type Node interface {
	QueryUtxos(types.Address) ([]tx.Utxo, error)
	QueryTip() (NodeTip, error)
	SubmitTx(tx.Transaction) error
}

type NodeTip struct {
	Epoch uint64
	Block uint64
	Slot  uint64
}
