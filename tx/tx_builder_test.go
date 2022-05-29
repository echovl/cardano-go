package tx

import (
	"testing"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/types"
)

var shelleyProtocol = &ProtocolParams{
	CoinsPerUTXOWord: 34482,
	MinFeeA:          44,
	MinFeeB:          155381,
}

func TestTXBuilder_AddFee(t *testing.T) {
	key := crypto.NewXPrv([]byte("receiver address"), "foo")
	receiver := types.NewEnterpriseAddress(key.PublicKey(), types.Testnet)

	type fields struct {
		tx       Transaction
		protocol *ProtocolParams
		inputs   []TransactionInput
		outputs  []TransactionOutput
		ttl      uint64
		fee      uint64
		vkeys    map[string]crypto.XPub
		pkeys    map[string]crypto.XPrv
	}
	tests := []struct {
		name      string
		fields    fields
		hasChange bool
		wantErr   bool
	}{
		{
			name: "input < output + fee",
			fields: fields{
				protocol: shelleyProtocol,
				inputs: []TransactionInput{
					{
						TxHash: [32]byte{},
						Index:  uint64(0),
						Amount: 200000,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver,
						Amount:  200000,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "input == output + fee",
			fields: fields{
				protocol: shelleyProtocol,
				inputs: []TransactionInput{
					{
						TxHash: [32]byte{},
						Index:  0,
						Amount: minUTXO(nil, shelleyProtocol) + 165501,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver,
						Amount:  minUTXO(nil, shelleyProtocol),
					},
				},
			},
		},
		{
			name: "input > output + fee AND change < min utxo value -> burned",
			fields: fields{
				protocol: shelleyProtocol,
				inputs: []TransactionInput{
					{
						TxHash: [32]byte{},
						Index:  0,
						Amount: 2*minUTXO(nil, shelleyProtocol) - 1,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver,
						Amount:  minUTXO(nil, shelleyProtocol),
					},
				},
			},
		},
		{
			name: "input > output + fee AND change > min utxo BUT change - change output fee < min utxo -> burned",
			fields: fields{
				protocol: shelleyProtocol,
				inputs: []TransactionInput{
					{
						TxHash: [32]byte{},
						Index:  0,
						Amount: 2*minUTXO(nil, shelleyProtocol) + 162685,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver,
						Amount:  minUTXO(nil, shelleyProtocol),
					},
				},
			},
		},
		{
			name: "input > output + fee AND change > min utxo AND change - change output fee > min utxo -> sent",
			fields: fields{
				protocol: shelleyProtocol,
				inputs: []TransactionInput{
					{
						TxHash: [32]byte{},
						Index:  0,
						Amount: 3 * minUTXO(nil, shelleyProtocol),
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver,
						Amount:  minUTXO(nil, shelleyProtocol),
					},
				},
			},
			hasChange: true,
		},
		{
			name: "take in account ttl -> burned",
			fields: fields{
				protocol: shelleyProtocol,
				inputs: []TransactionInput{
					{
						TxHash: [32]byte{},
						Index:  0,
						Amount: 2*minUTXO(nil, shelleyProtocol) + 164137,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver,
						Amount:  minUTXO(nil, shelleyProtocol),
					},
				},
				ttl: 100,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := crypto.NewXPrv([]byte("change address"), "foo")
			change := types.NewEnterpriseAddress(key.PublicKey(), types.Testnet)
			builder := NewTxBuilder(shelleyProtocol)
			builder.AddInputs(tt.fields.inputs...)
			builder.AddOutputs(tt.fields.outputs...)
			builder.SetTTL(tt.fields.ttl)
			builder.Sign(key)
			if err := builder.AddFee(change); err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("AddFee() error = %v, wantErr %v", err, tt.wantErr)
			}
			var totalIn types.Coin
			for _, input := range builder.tx.Body.Inputs {
				totalIn += input.Amount
			}
			var totalOut types.Coin
			for _, output := range builder.tx.Body.Outputs {
				totalOut += output.Amount
			}
			if got, want := builder.tx.Body.Fee+totalOut, totalIn; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			expectedReceiver := receiver
			if tt.hasChange {
				expectedReceiver = change
				if got, want := builder.tx.Body.Outputs[0].Amount, minUTXO(nil, builder.protocol); got < want {
					t.Errorf("got %v want greater than %v", got, want)
				}
			}
			firstOutputReceiver := builder.tx.Body.Outputs[0].Address
			if got, want := firstOutputReceiver.String(), expectedReceiver.String(); got != want {
				t.Errorf("got %v want %v", got, want)
			}
		})
	}
}
