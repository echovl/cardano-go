package tx

import (
	"errors"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/types"
)

const maxUint64 uint64 = 1<<64 - 1

type TXBuilder struct {
	tx       Transaction
	protocol types.ProtocolParams
	inputs   []TransactionInput
	outputs  []TransactionOutput
	ttl      int
	fee      types.Coin
	pkeys    []crypto.XPrv
}

// NewTxBuilder returns a new instance of TxBuilder.
func NewTxBuilder(protocol types.ProtocolParams) *TXBuilder {
	return &TXBuilder{
		protocol: protocol,
		pkeys:    []crypto.XPrv{},
	}
}

// AddInputs adds inputs to the transaction being builded.
func (builder *TXBuilder) AddInputs(inputs ...TransactionInput) {
	builder.inputs = append(builder.inputs, inputs...)
}

// AddOutputs adds outputs to the transaction being builded.
func (builder *TXBuilder) AddOutputs(outputs ...TransactionOutput) {
	builder.outputs = append(builder.outputs, outputs...)
}

// SetTtl sets the transaction's time to live.
func (builder *TXBuilder) SetTTL(ttl int) {
	builder.ttl = ttl
}

// SetFee sets the transactions's fee.
func (builder *TXBuilder) SetFee(fee types.Coin) {
	builder.fee = fee
}

// AddFee calculates the required fee for the transaction and adds
// an aditional output for the change if there is any.
// This assumes that the builder inputs and outputs are defined.
func (builder *TXBuilder) AddFee(changeAddr types.Address) error {
	inputAmount := types.Coin(0)
	for _, txIn := range builder.inputs {
		inputAmount += txIn.Amount
	}
	body := builder.buildBody()

	if err := body.addFee(inputAmount, changeAddr, builder.protocol); err != nil {
		return err
	}
	builder.outputs = body.Outputs
	builder.fee = body.Fee
	return nil
}

// Sign adds a signing key to create a signature for the witness set.
func (builder *TXBuilder) Sign(xsk crypto.XPrv) {
	builder.pkeys = append(builder.pkeys, xsk)
}

// Build creates a new transaction using the inputs, outputs and keys provided.
func (builder *TXBuilder) Build() (*Transaction, error) {
	if len(builder.pkeys) == 0 {
		return nil, errors.New("missing signing keys")
	}

	body := builder.buildBody()
	witnessSet := WitnessSet{}
	txHash, err := body.Hash()
	if err != nil {
		return nil, err
	}

	for _, pkey := range builder.pkeys {
		publicKey := pkey.PublicKey()[:32]
		signature := pkey.Sign(txHash[:])
		witness := VKeyWitness{VKey: publicKey, Signature: signature}
		witnessSet.VKeyWitnessSet = append(witnessSet.VKeyWitnessSet, witness)
	}

	tx := &Transaction{
		Body:          body,
		WitnessSet:    witnessSet,
		AuxiliaryData: nil,
		IsValid:       true,
	}

	return tx, nil
}

func (builder *TXBuilder) buildBody() TransactionBody {
	inputs := make([]TransactionInput, len(builder.inputs))
	for i, txInput := range builder.inputs {
		inputs[i] = TransactionInput{
			TxHash: txInput.TxHash,
			Index:  txInput.Index,
		}
	}

	return TransactionBody{
		Inputs:  inputs,
		Outputs: builder.outputs,
		Fee:     builder.fee,
		TTL:     types.NewUint64(uint64(builder.ttl)),
	}
}
