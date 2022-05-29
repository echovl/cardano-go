package tx

import (
	"errors"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/types"
	"github.com/fxamacker/cbor/v2"
	"golang.org/x/crypto/blake2b"
)

const maxUint64 uint64 = 1<<64 - 1

type TXBuilder struct {
	tx            Transaction
	protocol      *ProtocolParams
	inputs        []TransactionInput
	outputs       []TransactionOutput
	ttl           int
	fee           types.Coin
	auxiliaryData *AuxiliaryData
	pkeys         []crypto.XPrv
}

// NewTxBuilder returns a new instance of TxBuilder.
func NewTxBuilder(protocol *ProtocolParams) *TXBuilder {
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

func (builder *TXBuilder) AddAuxiliaryData(data *AuxiliaryData) {
	builder.auxiliaryData = data
}

// AddFee calculates the required fee for the transaction and adds
// an aditional output for the change if there is any.
// This assumes that the builder inputs and outputs are defined.
func (builder *TXBuilder) AddFee(changeAddr types.Address) error {
	inputAmount := types.Coin(0)
	for _, txIn := range builder.inputs {
		inputAmount += txIn.Amount
	}

	tx, err := builder.Build()
	if err != nil {
		return err
	}

	if err := tx.addFee(inputAmount, changeAddr, builder.protocol); err != nil {
		return err
	}

	builder.outputs = tx.Body.Outputs
	builder.fee = tx.Body.Fee

	return nil
}

// Sign adds signing keys to create signatures for the witness set.
func (builder *TXBuilder) Sign(xsk ...crypto.XPrv) {
	builder.pkeys = append(builder.pkeys, xsk...)
}

// Build creates a new transaction using the inputs, outputs and keys provided.
func (builder *TXBuilder) Build() (*Transaction, error) {
	if len(builder.pkeys) == 0 {
		return nil, errors.New("missing signing keys")
	}

	body, err := builder.buildBody()
	if err != nil {
		return nil, err
	}

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
		AuxiliaryData: builder.auxiliaryData,
		IsValid:       true,
	}

	return tx, nil
}

func (builder *TXBuilder) buildBody() (TransactionBody, error) {
	inputs := make([]TransactionInput, len(builder.inputs))
	for i, txInput := range builder.inputs {
		inputs[i] = TransactionInput{
			TxHash: txInput.TxHash,
			Index:  txInput.Index,
		}
	}

	body := TransactionBody{
		Inputs:  inputs,
		Outputs: builder.outputs,
		Fee:     builder.fee,
		TTL:     types.NewUint64(uint64(builder.ttl)),
	}

	if builder.auxiliaryData != nil {
		auxBytes, err := cbor.Marshal(builder.auxiliaryData)
		if err != nil {
			return body, err
		}
		auxHash := types.Hash32(blake2b.Sum256(auxBytes))
		body.AuxiliaryDataHash = &auxHash

	}

	return body, nil
}
