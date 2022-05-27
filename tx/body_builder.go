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

var ShelleyProtocol = types.ProtocolParams{
	MinimumUtxoValue: 1000000,
	MinFeeA:          44,
	MinFeeB:          155381,
}

func LiveTTL() uint64 {
	shelleyStart := time.Unix(shelleyStartTimestamp, 0)
	return uint64(shelleyStartSlot + time.Since(shelleyStart).Seconds())
}

type UTXO struct {
	TxHash  types.Hash32
	Spender types.Address
	Amount  types.Coin
	Index   uint64
}

type TXBodyBuilder struct {
	Protocol types.ProtocolParams
	TTL      uint64
}

func (builder TXBodyBuilder) Build(receiver types.Address, pickedUtxos []UTXO, amount types.Coin, change types.Address) (*TransactionBody, error) {
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
		Address: receiver.Bytes(),
		Amount:  amount,
	})

	body := TransactionBody{
		Inputs:  inputs,
		Outputs: outputs,
		TTL:     types.NewUint64(builder.ttl()),
	}
	if err := body.addFee(inputAmount, change, builder.protocol()); err != nil {
		return nil, err
	}

	return &body, nil
}

func (builder TXBodyBuilder) ttl() uint64 {
	if builder.TTL == 0 {
		return LiveTTL() + slotMargin
	}
	return builder.TTL
}

func (builder TXBodyBuilder) protocol() types.ProtocolParams {
	if builder.Protocol == (types.ProtocolParams{}) {
		return ShelleyProtocol
	}
	return builder.Protocol
}
