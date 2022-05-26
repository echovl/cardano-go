package tx

import (
	"encoding/hex"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/types"
	"golang.org/x/crypto/blake2b"
)

const maxUint64 uint64 = 1<<64 - 1

type TXBuilderInput struct {
	input  TransactionInput
	amount types.Coin
}

type TXBuilderOutput struct {
	address types.Address
	amount  types.Coin
}

type TXBuilder struct {
	tx       Transaction
	protocol types.ProtocolParams
	inputs   []TXBuilderInput
	outputs  []TransactionOutput
	ttl      uint64
	fee      types.Coin
	vkeys    map[string]crypto.ExtendedVerificationKey
	pkeys    map[string]crypto.ExtendedSigningKey
}

func NewTxBuilder(protocol types.ProtocolParams) *TXBuilder {
	return &TXBuilder{
		protocol: protocol,
		vkeys:    map[string]crypto.ExtendedVerificationKey{},
		pkeys:    map[string]crypto.ExtendedSigningKey{},
	}
}

func (builder *TXBuilder) AddInput(xvk crypto.ExtendedVerificationKey, txHash types.Hash32, index uint64, amount types.Coin) {
	input := TXBuilderInput{input: TransactionInput{TxHash: txHash, Index: index}, amount: amount}
	builder.inputs = append(builder.inputs, input)

	vkeyHashBytes := blake2b.Sum256(xvk)
	vkeyHashString := hex.EncodeToString(vkeyHashBytes[:])
	builder.vkeys[vkeyHashString] = xvk
}

func (builder *TXBuilder) AddInputWithoutSig(txHash types.Hash32, index uint64, amount types.Coin) {
	input := TXBuilderInput{input: TransactionInput{TxHash: txHash, Index: index}, amount: amount}
	builder.inputs = append(builder.inputs, input)
}

func (builder *TXBuilder) AddOutput(address types.Address, amount types.Coin) {
	output := TransactionOutput{Address: address.Bytes(), Amount: amount}
	builder.outputs = append(builder.outputs, output)
}

func (builder *TXBuilder) SetTtl(ttl uint64) {
	builder.ttl = ttl
}

func (builder *TXBuilder) SetFee(fee types.Coin) {
	builder.fee = fee
}

// This assumes that the builder inputs and outputs are defined
func (builder *TXBuilder) AddFee(address types.Address) error {
	inputAmount := types.Coin(0)
	for _, txIn := range builder.inputs {
		inputAmount += txIn.amount
	}
	body := builder.buildBody()

	if err := body.addFee(inputAmount, address, builder.protocol); err != nil {
		return err
	}
	builder.outputs = body.Outputs
	builder.fee = body.Fee
	return nil
}

func (builder *TXBuilder) Sign(xsk crypto.ExtendedSigningKey) {
	pkeyHashBytes := blake2b.Sum256(xsk)
	pkeyHashString := hex.EncodeToString(pkeyHashBytes[:])
	builder.pkeys[pkeyHashString] = xsk
}

func (builder *TXBuilder) Build() Transaction {
	if len(builder.pkeys) != len(builder.vkeys) {
		panic("missing signatures")
	}

	body := builder.buildBody()
	witnessSet := WitnessSet{}
	txHash := blake2b.Sum256(body.Bytes())
	for _, pkey := range builder.pkeys {
		publicKey := pkey.ExtendedVerificationKey()[:32]
		signature := pkey.Sign(txHash[:])
		witness := VKeyWitness{VKey: publicKey, Signature: signature}
		witnessSet.VKeyWitnessSet = append(witnessSet.VKeyWitnessSet, witness)
	}

	return Transaction{
		Body:          body,
		WitnessSet:    witnessSet,
		AuxiliaryData: nil,
		IsValid:       true,
	}
}

func (builder *TXBuilder) buildBody() TransactionBody {
	inputs := make([]TransactionInput, len(builder.inputs))
	for i, txInput := range builder.inputs {
		inputs[i] = TransactionInput{
			TxHash: txInput.input.TxHash,
			Index:  txInput.input.Index,
		}
	}

	return TransactionBody{
		Inputs:  inputs,
		Outputs: builder.outputs,
		Fee:     builder.fee,
		TTL:     types.NewUint64(builder.ttl),
	}
}
