package tx

import (
	"testing"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/types"
)

var shelleyProtocol = &ProtocolParams{
	MinUTXO: 1000000,
	MinFeeA: 44,
	MinFeeB: 155381,
}

func TestTXBuilder_AddFee(t *testing.T) {
	key := crypto.NewXPrv([]byte("receiver address"), "foo")
	receiver := types.NewEnterpriseAddress(key.PublicKey(), types.Testnet)

	type fields struct {
		tx       Transaction
		protocol *ProtocolParams
		inputs   []TransactionInput
		outputs  []TransactionOutput
		ttl      int
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
						Amount: shelleyProtocol.MinUTXO + 1162729,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver,
						Amount:  shelleyProtocol.MinUTXO,
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
						Amount: 2*shelleyProtocol.MinUTXO - 1,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver,
						Amount:  shelleyProtocol.MinUTXO,
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
						Amount: 2*shelleyProtocol.MinUTXO + 162685,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver,
						Amount:  shelleyProtocol.MinUTXO,
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
						Amount: 3 * shelleyProtocol.MinUTXO,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver,
						Amount:  shelleyProtocol.MinUTXO,
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
						Amount: 2*shelleyProtocol.MinUTXO + 164137,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver,
						Amount:  shelleyProtocol.MinUTXO,
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
			builder := &TXBuilder{
				protocol: tt.fields.protocol,
				inputs:   tt.fields.inputs,
				outputs:  tt.fields.outputs,
				ttl:      tt.fields.ttl,
				pkeys:    []crypto.XPrv{key},
			}
			if err := builder.AddFee(change); err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("AddFee() error = %v, wantErr %v", err, tt.wantErr)
			}
			var totalIn types.Coin
			for _, input := range builder.inputs {
				totalIn += input.Amount
			}
			var totalOut types.Coin
			for _, output := range builder.outputs {
				totalOut += output.Amount
			}
			if got, want := builder.fee+totalOut, totalIn; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			expectedReceiver := receiver
			if tt.hasChange {
				expectedReceiver = change
				if got, want := builder.outputs[0].Amount, builder.protocol.MinUTXO; got < want {
					t.Errorf("got %v want greater than %v", got, want)
				}
			}
			firstOutputReceiver := builder.outputs[0].Address
			if got, want := firstOutputReceiver.String(), expectedReceiver.String(); got != want {
				t.Errorf("got %v want %v", got, want)
			}
		})
	}
}
