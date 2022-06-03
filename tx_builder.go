package cardano

import (
	"errors"
	"fmt"

	"github.com/echovl/cardano-go/crypto"
	"golang.org/x/crypto/blake2b"
)

type TxBuilder struct {
	tx       *Tx
	protocol *ProtocolParams
	pkeys    []crypto.PrvKey
}

// NewTxBuilder returns a new instance of TxBuilder.
func NewTxBuilder(protocol *ProtocolParams) *TxBuilder {
	return &TxBuilder{
		protocol: protocol,
		pkeys:    []crypto.PrvKey{},
		tx: &Tx{
			IsValid: true,
		},
	}
}

// AddInputs adds inputs to the transaction.
func (tb *TxBuilder) AddInputs(inputs ...*TxInput) {
	tb.tx.Body.Inputs = append(tb.tx.Body.Inputs, inputs...)
}

// AddOutputs adds outputs to the transaction.
func (tb *TxBuilder) AddOutputs(outputs ...*TxOutput) {
	tb.tx.Body.Outputs = append(tb.tx.Body.Outputs, outputs...)
}

// SetTtl sets the transaction's time to live.
func (tb *TxBuilder) SetTTL(ttl uint64) {
	tb.tx.Body.TTL = NewUint64(ttl)
}

// SetFee sets the transactions's fee.
func (tb *TxBuilder) SetFee(fee Coin) {
	tb.tx.Body.Fee = fee
}

// AddAuxiliaryData adds auxiliary data to the transaction.
func (tb *TxBuilder) AddAuxiliaryData(data *AuxiliaryData) {
	tb.tx.AuxiliaryData = data
}

// AddCertificate adds a certificate to the transaction.
func (tb *TxBuilder) AddCertificate(cert Certificate) {
	tb.tx.Body.Certificates = append(tb.tx.Body.Certificates, cert)
}

// AddNativeScript adds a native script to the transaction.
func (tb *TxBuilder) AddNativeScript(script NativeScript) {
	tb.tx.WitnessSet.Scripts = append(tb.tx.WitnessSet.Scripts, script)
}

// Mint adds a new multiasset to mint.
func (tb *TxBuilder) Mint(asset *Mint) {
	tb.tx.Body.Mint = asset
}

// AddChangeIfNeeded calculates the required fee for the transaction and adds
// an aditional output for the change if there is any.
// This assumes that the inputs-outputs are defined and signing keys are present.
func (tb *TxBuilder) AddChangeIfNeeded(changeAddr Address) error {
	inputAmount, outputAmount := tb.calculateAmounts()
	totalDeposits := tb.totalDeposits()

	// Set a temporary realistic fee in order to serialize a valid transaction
	tb.tx.Body.Fee = 200000
	if _, err := tb.build(); err != nil {
		return err
	}

	if tb.tx.Body.Mint != nil {
		inputAmount = inputAmount.Add(NewValueWithAssets(0, tb.tx.Body.Mint.MultiAsset()))
	}

	minFee := tb.calculateMinFee()
	outputAmount = outputAmount.Add(NewValue(minFee + totalDeposits))

	if inputOutputCmp := inputAmount.Cmp(outputAmount); inputOutputCmp == -1 {
		return fmt.Errorf(
			"insuficient input in transaction, got %v want atleast %v",
			inputAmount,
			outputAmount,
		)
	} else if inputOutputCmp == 0 {
		tb.tx.Body.Fee = minFee
		return nil
	} else if inputOutputCmp == 2 {
		return fmt.Errorf(
			"inputs and outputs with different assets, input %v and output %v",
			inputAmount,
			outputAmount,
		)
	}

	// Construct change output
	changeAmount := inputAmount.Sub(outputAmount)
	changeOutput := NewTxOutput(changeAddr, changeAmount)

	changeMinUTXO := minUTXO(changeOutput, tb.protocol)
	if changeAmount.Coin < changeMinUTXO {
		if changeAmount.OnlyCoin() {
			tb.tx.Body.Fee = minFee + changeAmount.Coin // burn change
			return nil
		}
		return fmt.Errorf(
			"insuficient input for change output with multiassets, got %v want %v",
			inputAmount.Coin,
			inputAmount.Coin+changeMinUTXO-changeAmount.Coin,
		)
	}

	tb.tx.Body.Outputs = append([]*TxOutput{changeOutput}, tb.tx.Body.Outputs...)

	newMinFee := tb.calculateMinFee()
	changeAmount.Coin = changeAmount.Coin + minFee - newMinFee
	if changeAmount.Coin < changeMinUTXO {
		if changeAmount.OnlyCoin() {
			tb.tx.Body.Fee = newMinFee + changeAmount.Coin // burn change
			tb.tx.Body.Outputs = tb.tx.Body.Outputs[1:]    // remove change output
			return nil
		}
		return fmt.Errorf(
			"insuficient input for change output with multiassets, got %v want %v",
			inputAmount.Coin,
			changeMinUTXO,
		)
	}

	tb.tx.Body.Fee = newMinFee

	return nil
}

func (tb *TxBuilder) calculateAmounts() (*Value, *Value) {
	input, output := NewValue(0), NewValue(0)
	for _, in := range tb.tx.Body.Inputs {
		input = input.Add(in.Amount)
	}
	for _, out := range tb.tx.Body.Outputs {
		output = output.Add(out.Amount)
	}
	return input, output
}

func (tb *TxBuilder) totalDeposits() Coin {
	certs := tb.tx.Body.Certificates
	var deposit Coin
	if len(certs) != 0 {
		for _, cert := range certs {
			if cert.Type == StakeRegistration {
				deposit += tb.protocol.KeyDeposit
			}
		}
	}
	return deposit
}

// MinFee computes the minimal fee required for the transaction.
// This assumes that the inputs-outputs are defined and signing keys are present.
func (tb *TxBuilder) MinFee() (Coin, error) {
	// Set a temporary realistic fee in order to serialize a valid transaction
	tb.tx.Body.Fee = 200000
	if _, err := tb.build(); err != nil {
		return 0, err
	}
	minFee := tb.calculateMinFee()
	return minFee, nil
}

// CalculateFee computes the minimal fee required for the transaction.
func (tb *TxBuilder) calculateMinFee() Coin {
	txBytes := tb.tx.Bytes()
	txLength := uint64(len(txBytes))
	return tb.protocol.MinFeeA*Coin(txLength) + tb.protocol.MinFeeB
}

// Sign adds signing keys to create signatures for the witness set.
func (tb *TxBuilder) Sign(privateKeys ...crypto.PrvKey) error {
	tb.pkeys = append(tb.pkeys, privateKeys...)
	return nil
}

// Build creates a new transaction using the inputs, outputs and keys provided.
func (tb *TxBuilder) Build() (*Tx, error) {
	inputAmount, outputAmount := tb.calculateAmounts()
	outputAmount = outputAmount.Add(NewValue(tb.tx.Body.Fee)).Add(NewValue(tb.totalDeposits()))

	if tb.tx.Body.Mint != nil {
		inputAmount = inputAmount.Add(NewValueWithAssets(0, tb.tx.Body.Mint.MultiAsset()))
	}

	if inputOutputCmp := outputAmount.Cmp(inputAmount); inputOutputCmp == 1 {
		return nil, fmt.Errorf(
			"insuficient input in transaction, got %v want %v",
			inputAmount,
			outputAmount,
		)
	} else if inputOutputCmp == -1 {
		return nil, fmt.Errorf(
			"fee too small, got %v want %v",
			tb.tx.Body.Fee,
			inputAmount.Sub(outputAmount),
		)
	} else if inputOutputCmp == 2 {
		return nil, fmt.Errorf(
			"inputs and outputs with different assets, input %v and output %v",
			inputAmount,
			outputAmount,
		)
	}

	return tb.build()
}

func (tb *TxBuilder) build() (*Tx, error) {
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
		publicKey := pkey.PubKey()
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
		auxHash := blake2b.Sum256(auxBytes)
		auxHash32 := Hash32(auxHash[:])
		tb.tx.Body.AuxiliaryDataHash = &auxHash32
	}
	return nil
}
