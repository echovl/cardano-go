package node

import (
	"github.com/echovl/cardano-go/tx"
	"github.com/echovl/cardano-go/types"
)

const (
	protocolMagic = 1097911063
)

type Node interface {
	UTXOs(types.Address) ([]tx.Utxo, error)
	Tip() (NodeTip, error)
	SubmitTx(tx.Transaction) error
	Network() types.Network
}

type NodeTip struct {
	Epoch uint64
	Block uint64
	Slot  uint64
}
