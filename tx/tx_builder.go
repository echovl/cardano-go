package tx

import (
	"errors"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/types"
	"golang.org/x/crypto/blake2b"
)

const maxUint64 uint64 = 1<<64 - 1

type TXBuilder struct {
	tx       *Transaction
	protocol *ProtocolParams
	pkeys    []crypto.XPrv
}

// NewTxBuilder returns a new instance of TxBuilder.
func NewTxBuilder(protocol *ProtocolParams) *TXBuilder {
	return &TXBuilder{
		protocol: protocol,
		pkeys:    []crypto.XPrv{},
		tx: &Transaction{
			IsValid: true,
		},
	}
}

// AddInputs adds inputs to the transaction being builded.
func (tb *TXBuilder) AddInputs(inputs ...TransactionInput) {
	tb.tx.Body.Inputs = append(tb.tx.Body.Inputs, inputs...)
}

// AddOutputs adds outputs to the transaction being builded.
func (tb *TXBuilder) AddOutputs(outputs ...TransactionOutput) {
	tb.tx.Body.Outputs = append(tb.tx.Body.Outputs, outputs...)
}

// SetTtl sets the transaction's time to live.
func (tb *TXBuilder) SetTTL(ttl uint64) {
	tb.tx.Body.TTL = types.NewUint64(ttl)
}

// SetFee sets the transactions's fee.
func (tb *TXBuilder) SetFee(fee types.Coin) {
	tb.tx.Body.Fee = fee
}

func (tb *TXBuilder) AddAuxiliaryData(data *AuxiliaryData) {
	tb.tx.AuxiliaryData = data
}

// AddFee calculates the required fee for the transaction and adds
// an aditional output for the change if there is any.
// This assumes that the tb inputs and outputs are defined.
func (tb *TXBuilder) AddFee(changeAddr types.Address) error {
	inputAmount := types.Coin(0)
	for _, txIn := range tb.tx.Body.Inputs {
		inputAmount += txIn.Amount
	}

	tx, err := tb.Build()
	if err != nil {
		return err
	}

	if err := tx.addFee(inputAmount, changeAddr, tb.protocol); err != nil {
		return err
	}

	return nil
}

// Sign adds signing keys to create signatures for the witness set.
func (tb *TXBuilder) Sign(xsk ...crypto.XPrv) {
	tb.pkeys = append(tb.pkeys, xsk...)
}

// Build creates a new transaction using the inputs, outputs and keys provided.
func (tb *TXBuilder) Build() (*Transaction, error) {
	if len(tb.pkeys) == 0 {
		return nil, errors.New("missing signing keys")
	}

	if err := tb.buildBody(); err != nil {
		return nil, err
	}

	txHash, err := tb.tx.Hash()
	if err != nil {
		return nil, err
	}

	vkeyWitnsessSet := make([]VKeyWitness, len(tb.pkeys))
	for i, pkey := range tb.pkeys {
		publicKey := pkey.PublicKey()[:32]
		signature := pkey.Sign(txHash[:])
		witness := VKeyWitness{VKey: publicKey, Signature: signature}
		vkeyWitnsessSet[i] = witness
	}
	tb.tx.WitnessSet.VKeyWitnessSet = vkeyWitnsessSet

	return tb.tx, nil
}

func (tb *TXBuilder) buildBody() error {
	if tb.tx.AuxiliaryData != nil {
		auxBytes, err := cborEnc.Marshal(tb.tx.AuxiliaryData)
		if err != nil {
			return err
		}
		auxHash := types.Hash32(blake2b.Sum256(auxBytes))
		tb.tx.Body.AuxiliaryDataHash = &auxHash
	}
	return nil
}
