package cardano

const (
	ProtocolMagic = 1097911063
)

// Node is the interface required for a Cardano backend/node.
// A backend/node is used to interact with the Cardano Blockchain,
// sending transactions and fetching state.
type Node interface {
	// UTxOs returns a list of unspent transaction outputs for a given address
	UTxOs(Address) ([]UTxO, error)

	// Tip returns the node's current tip
	Tip() (*NodeTip, error)

	// SubmitTx submits a transaction to the node using cbor encoding
	SubmitTx(*Tx) (*Hash32, error)

	// ProtocolParams returns the Node's Protocol Parameters
	ProtocolParams() (*ProtocolParams, error)

	// Network returns the node's current network type
	Network() Network
}

type NodeTip struct {
	Block uint64
	Epoch uint64
	Slot  uint64
}
