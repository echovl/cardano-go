package tx

import (
	"time"

	"github.com/echovl/cardano-go/types"
)

const (
	shelleyStartTimestamp = 1596491091
	shelleyStartSlot      = 4924800
	slotMargin            = 1200
)

var shelleyProtocol = types.ProtocolParams{
	MinimumUtxoValue: 1000000,
	MinFeeA:          44,
	MinFeeB:          155381,
}

func liveTTL() int {
	shelleyStart := time.Unix(shelleyStartTimestamp, 0)
	return int(shelleyStartSlot + time.Since(shelleyStart).Seconds())
}

type txBodyBuilder struct {
	Protocol types.ProtocolParams
	TTL      int
}

func (builder txBodyBuilder) build(receiver types.Address, pickedUtxos []UTXO, amount types.Coin, change types.Address) (*TransactionBody, error) {
	var inputAmount types.Coin
	var inputs []TransactionInput
	for _, utxo := range pickedUtxos {
		inputs = append(inputs, TransactionInput{
			TxHash: utxo.TxHash,
			Index:  utxo.Index,
		})
		inputAmount += utxo.Amount
	}

	var outputs []TransactionOutput
	outputs = append(outputs, TransactionOutput{
		Address: receiver,
		Amount:  amount,
	})

	body := TransactionBody{
		Inputs:  inputs,
		Outputs: outputs,
		TTL:     types.NewUint64(uint64(builder.ttl())),
	}
	if err := body.addFee(inputAmount, change, builder.protocol()); err != nil {
		return nil, err
	}

	return &body, nil
}

func (builder txBodyBuilder) ttl() int {
	if builder.TTL == 0 {
		return liveTTL() + slotMargin
	}
	return builder.TTL
}

func (builder txBodyBuilder) protocol() types.ProtocolParams {
	if builder.Protocol == (types.ProtocolParams{}) {
		return shelleyProtocol
	}
	return builder.Protocol
}
