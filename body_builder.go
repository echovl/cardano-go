package cardano

import "time"

const (
	shelleyStartTimestamp = 1596491091
	shelleyStartSlot      = 4924800
	slotMargin            = 1200
)

var ShelleyProtocol = ProtocolParams{
	MinimumUtxoValue: 1000000,
	MinFeeA:          44,
	MinFeeB:          155381,
}

func LiveTTL() uint64 {
	shelleyStart := time.Unix(shelleyStartTimestamp, 0)
	return uint64(shelleyStartSlot + time.Since(shelleyStart).Seconds())
}

type TXBodyBuilder struct {
	Protocol ProtocolParams
	TTL      uint64
}

func (builder TXBodyBuilder) Build(receiver Address, pickedUtxos []Utxo, amount uint64, change Address) (*TransactionBody, error) {
	var inputAmount uint64
	var inputs []TransactionInput
	for _, utxo := range pickedUtxos {
		inputs = append(inputs, TransactionInput{
			ID:    utxo.TxId.Bytes(),
			Index: utxo.Index,
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
		Ttl:     builder.ttl(),
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

func (builder TXBodyBuilder) protocol() ProtocolParams {
	if builder.Protocol == (ProtocolParams{}) {
		return ShelleyProtocol
	}
	return builder.Protocol
}
