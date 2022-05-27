package node

import (
	"github.com/echovl/cardano-go/tx"
	"github.com/echovl/cardano-go/types"
)

const (
	ProtocolMagic = 1097911063
)

// Node is the interface required for a Cardano backend/node.
// A backend/node is used to interact with the Cardano Blockchain,
// sending transactions and fetching state.
type Node interface {
	// UTXOs returns a list of unspent transaction outputs for a given address
	UTXOs(types.Address) ([]tx.UTXO, error)

	// Tip returns the node's current tip
	Tip() (*NodeTip, error)

	// SubmitTx submits a transaction to the node using cbor encoding
	SubmitTx(*tx.Transaction) (*types.Hash32, error)

	// Network returns the node's current network type
	Network() types.Network
}

type NodeTip struct {
	Block int
	Epoch int
	Slot  int
}
