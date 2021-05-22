package wallet

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/echovl/cardano-wallet/crypto"
	"golang.org/x/crypto/blake2b"
)

const maxUint64 uint64 = 1<<64 - 1

type TxId string

func (id TxId) Bytes() []byte {
	bytes, err := hex.DecodeString(string(id))
	if err != nil {
		panic(err)
	}

	return bytes
}

type TxBuilderInput struct {
	input  TransactionInput
	amount uint64
}

type TxBuilderOutput struct {
	address Address
	amount  uint64
}

type TxBuilder struct {
	transaction    Transaction
	protocolParams ProtocolParams
	inputs         []TxBuilderInput
	outputs        []TxBuilderOutput
	ttl            uint64
	fee            uint64
	vkeys          []crypto.XVerificationKey
	pkeys          []crypto.XSigningKey
}

func NewTxBuilder(protocolParams ProtocolParams) *TxBuilder {
	return &TxBuilder{protocolParams: protocolParams}
}

func (builder *TxBuilder) AddInput(xvk crypto.XVerificationKey, txId TxId, index, amount uint64) {
	input := TxBuilderInput{input: TransactionInput{ID: txId.Bytes(), Index: index}, amount: amount}
	builder.inputs = append(builder.inputs, input)
	builder.vkeys = append(builder.vkeys, xvk)
}

func (builder *TxBuilder) AddOutput(address Address, amount uint64) {
	output := TxBuilderOutput{address: address, amount: amount}
	builder.outputs = append(builder.outputs, output)
}

func (builder *TxBuilder) SetTtl(ttl uint64) {
	builder.ttl = ttl
}

func (builder *TxBuilder) SetFee(fee uint64) {
	builder.fee = fee
}

// This assumes that the builder inputs and outputs are defined
func (builder *TxBuilder) AddFee(address Address) error {
	// Set a temporary fee in order to serialize a valid transaction
	builder.SetFee(maxUint64)
	minFee := builder.calculateMinFee()
	inputAmount := uint64(0)
	for _, txIn := range builder.inputs {
		inputAmount += txIn.amount
	}

	outputAmount := uint64(0)
	for _, txOut := range builder.outputs {
		outputAmount += txOut.amount
	}
	outputWithFeeAmount := outputAmount + minFee

	if inputAmount > outputWithFeeAmount {
		minAda := builder.protocolParams.MinimumUtxoValue
		change := inputAmount - outputWithFeeAmount
		if change > minAda {
			feeChange := builder.feeForOuput(address, change)
			newFee := minFee + feeChange
			change = inputAmount - (outputAmount + newFee)

			if change > minAda {
				builder.AddOutput(address, change)
				builder.SetFee(newFee)
				log.Println("Adding change output")
			} else {
				builder.SetFee(minFee)
				log.Println("Burning remaining change")
			}
		} else {
			builder.SetFee(minFee)
			log.Println("Burning remaining change")
		}

	} else if inputAmount == outputWithFeeAmount {
		builder.SetFee(minFee)
		log.Println("No remaining change")
	} else {
		return fmt.Errorf("Insuficient input in transaction")
	}

	return nil
}

func (builder *TxBuilder) calculateMinFee() uint64 {
	fakeXSigningKey := crypto.GenerateMasterKey([]byte{
		0x0c, 0xcb, 0x74, 0xf3, 0x6b, 0x7d, 0xa1, 0x64, 0x9a, 0x81, 0x44, 0x67, 0x55, 0x22, 0xd4, 0xd8, 0x09, 0x7c, 0x64, 0x12,
	}, "")

	body := builder.buildBody()

	witnessSet := TransactionWitnessSet{}
	for _, vkey := range builder.vkeys {
		vkeyWitness := VKeyWitness{VKey: fakeXSigningKey.XVerificationKey()[:32], Signature: fakeXSigningKey.Sign(vkey)}
		witnessSet.VKeyWitnessSet = append(witnessSet.VKeyWitnessSet, vkeyWitness)
	}

	transaction := Transaction{
		Body:       body,
		WitnessSet: witnessSet,
		Metadata:   nil,
	}

	return builder.calculateFee(&transaction)
}

func (builder *TxBuilder) feeForOuput(address Address, amount uint64) uint64 {
	builderCpy := *builder

	// We don't care about the fee impact on the tx size since we are
	// calculating the fee difference
	builderCpy.SetFee(0)

	feeBefore := builderCpy.calculateMinFee()
	builderCpy.AddOutput(address, amount)
	feeAfter := builderCpy.calculateMinFee()

	return feeAfter - feeBefore
}

func (builder *TxBuilder) calculateFee(tx *Transaction) uint64 {
	txBytes := tx.Bytes()
	txLength := uint64(len(txBytes))
	return builder.protocolParams.MinFeeA*txLength + builder.protocolParams.MinFeeB
}

func (builder *TxBuilder) Sign(xsk crypto.XSigningKey) {
	builder.pkeys = append(builder.pkeys, xsk)
}

func (builder *TxBuilder) Build() Transaction {
	if len(builder.pkeys) != len(builder.vkeys) {
		panic("missing signatures")
	}

	body := builder.buildBody()
	witnessSet := TransactionWitnessSet{}
	for _, pkey := range builder.pkeys {
		txHash := blake2b.Sum256(body.Bytes())
		publicKey := pkey.XVerificationKey()[:32]
		signature := pkey.Sign(txHash[:])
		vkeyWitness := VKeyWitness{VKey: publicKey, Signature: signature}

		witnessSet.VKeyWitnessSet = append(witnessSet.VKeyWitnessSet, vkeyWitness)
	}

	return Transaction{Body: body, WitnessSet: witnessSet, Metadata: nil}
}

func (builder *TxBuilder) buildBody() TransactionBody {
	inputs := make([]TransactionInput, len(builder.inputs))
	for i, txInput := range builder.inputs {
		inputs[i] = TransactionInput{
			ID:    txInput.input.ID,
			Index: txInput.input.Index,
		}
	}

	outputs := make([]TransactionOutput, len(builder.outputs))
	for i, txOutput := range builder.outputs {
		outputs[i] = TransactionOutput{
			Address: txOutput.address.Bytes(),
			Amount:  txOutput.amount,
		}
	}
	return TransactionBody{
		Inputs:  inputs,
		Outputs: outputs,
		Fee:     builder.fee,
		Ttl:     builder.ttl,
	}
}

func pretty(v interface{}) string {
	bytes, _ := json.MarshalIndent(v, "", "  ")
	return string(bytes)
}
