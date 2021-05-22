package wallet

type Utxo struct {
	Address Address
	TxId    TxId
	Amount  uint64
	Index   uint64
}

type NodeTip struct {
	Epoch uint64
	Block uint64
	Slot  uint64
}

type CardanoNode interface {
	QueryUtxos(Address) ([]Utxo, error)
	QueryTip() (NodeTip, error)
	SubmitTx(Transaction) error
}
