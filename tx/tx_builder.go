package tx

import (
	"errors"
	"fmt"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/types"
	"golang.org/x/crypto/blake2b"
)

const maxUint64 uint64 = 1<<64 - 1

type TxBuilder struct {
	tx       *Tx
	protocol *ProtocolParams
	pkeys    []crypto.XPrv
}

// NewTxBuilder returns a new instance of TxBuilder.
func NewTxBuilder(protocol *ProtocolParams) *TxBuilder {
	return &TxBuilder{
		protocol: protocol,
		pkeys:    []crypto.XPrv{},
		tx: &Tx{
			IsValid: true,
		},
	}
}

// AddInputs adds inputs to the transaction being builded.
func (tb *TxBuilder) AddInputs(inputs ...*TxInput) {
	tb.tx.Body.Inputs = append(tb.tx.Body.Inputs, inputs...)
}

// AddOutputs adds outputs to the transaction being builded.
func (tb *TxBuilder) AddOutputs(outputs ...*TxOutput) {
	tb.tx.Body.Outputs = append(tb.tx.Body.Outputs, outputs...)
}

// SetTtl sets the transaction's time to live.
func (tb *TxBuilder) SetTTL(ttl uint64) {
	tb.tx.Body.TTL = types.NewUint64(ttl)
}

// SetFee sets the transactions's fee.
func (tb *TxBuilder) SetFee(fee types.Coin) {
	tb.tx.Body.Fee = fee
}

func (tb *TxBuilder) AddAuxiliaryData(data *AuxiliaryData) {
	tb.tx.AuxiliaryData = data
}

// AddChangeIfNeeded calculates the required fee for the transaction and adds
// an aditional output for the change if there is any.
// This assumes that the tb inputs and outputs are defined.
func (tb *TxBuilder) AddChangeIfNeeded(changeAddr string) error {
	addr, err := types.NewAddress(changeAddr)
	if err != nil {
		return err
	}
	inputAmount := types.Coin(0)
	for _, txIn := range tb.tx.Body.Inputs {
		inputAmount += txIn.Amount
	}

	if _, err := tb.Build(); err != nil {
		return err
	}

	if err := tb.addFee(inputAmount, addr, tb.protocol); err != nil {
		return err
	}

	return nil
}

func (builder *TxBuilder) addFee(inputAmount types.Coin, changeAddress types.Address, protocol *ProtocolParams) error {
	// Set a temporary realistic fee in order to serialize a valid transaction
	builder.tx.Body.Fee = 200000

	minFee := builder.CalculateFee(protocol)

	outputAmount := types.Coin(0)
	for _, txOut := range builder.tx.Body.Outputs {
		outputAmount += txOut.Amount
	}
	outputWithFeeAmount := outputAmount + minFee

	if inputAmount < outputWithFeeAmount {
		return fmt.Errorf(
			"insuficient input in transaction, got %v want atleast %v",
			inputAmount,
			outputWithFeeAmount,
		)
	}

	if inputAmount == outputWithFeeAmount {
		builder.tx.Body.Fee = minFee
		return nil
	}

	change := inputAmount - outputWithFeeAmount
	changeOutput := &TxOutput{
		Address: changeAddress,
		Amount:  change,
	}
	changeMinUTXO := minUTXO(changeOutput, protocol)
	if change < changeMinUTXO {
		builder.tx.Body.Fee = minFee + change // burn change
		return nil
	}

	builder.tx.Body.Outputs = append([]*TxOutput{changeOutput}, builder.tx.Body.Outputs...)

	newMinFee := builder.CalculateFee(protocol)
	if change+minFee-newMinFee < changeMinUTXO {
		builder.tx.Body.Fee = minFee + change                 // burn change
		builder.tx.Body.Outputs = builder.tx.Body.Outputs[1:] // remove change output
		return nil
	}

	builder.tx.Body.Outputs[0].Amount = change + minFee - newMinFee
	builder.tx.Body.Fee = newMinFee

	return nil
}

// CalculateFee computes the minimal fee required for the transaction.
func (builder *TxBuilder) CalculateFee(protocol *ProtocolParams) types.Coin {
	txBytes := builder.tx.Bytes()
	txLength := uint64(len(txBytes))
	return protocol.MinFeeA*types.Coin(txLength) + protocol.MinFeeB
}

// Sign adds signing keys to create signatures for the witness set.
func (tb *TxBuilder) Sign(keys ...string) error {
	for _, key := range keys {
		xprv, err := crypto.NewXPrvFromBech32(key)
		if err != nil {
			return err
		}
		tb.pkeys = append(tb.pkeys, xprv)
	}
	return nil
}

// Build creates a new transaction using the inputs, outputs and keys provided.
func (tb *TxBuilder) Build() (*Tx, error) {
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

func (tb *TxBuilder) buildBody() error {
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
