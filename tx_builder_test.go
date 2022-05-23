package cardano

import (
	"github.com/tclairet/cardano-go/crypto"
	"testing"
)

func TestTXBuilder_AddFee(t *testing.T) {
	key := crypto.NewExtendedSigningKey([]byte("receiver address"), "foo")
	receiver := NewEnterpriseAddress(key.ExtendedVerificationKey(), Testnet)
	type fields struct {
		tx       Transaction
		protocol ProtocolParams
		inputs   []TXBuilderInput
		outputs  []TransactionOutput
		ttl      uint64
		fee      uint64
		vkeys    map[string]crypto.ExtendedVerificationKey
		pkeys    map[string]crypto.ExtendedSigningKey
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
				protocol: ShelleyProtocol,
				inputs: []TXBuilderInput{
					{
						input: TransactionInput{
							ID:    []byte("input 0"),
							Index: 0,
						},
						amount: 200000,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver.Bytes(),
						Amount:  200000,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "input == output + fee",
			fields: fields{
				protocol: ShelleyProtocol,
				inputs: []TXBuilderInput{
					{
						input: TransactionInput{
							ID:    []byte("input 0"),
							Index: 0,
						},
						amount: ShelleyProtocol.MinimumUtxoValue + 162685,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver.Bytes(),
						Amount:  ShelleyProtocol.MinimumUtxoValue,
					},
				},
			},
		},
		{
			name: "input > output + fee AND change < min utxo value -> burned",
			fields: fields{
				protocol: ShelleyProtocol,
				inputs: []TXBuilderInput{
					{
						input: TransactionInput{
							ID:    []byte("input 0"),
							Index: 0,
						},
						amount: 2*ShelleyProtocol.MinimumUtxoValue - 1,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver.Bytes(),
						Amount:  ShelleyProtocol.MinimumUtxoValue,
					},
				},
			},
		},
		{
			name: "input > output + fee AND change > min utxo BUT change - change output fee < min utxo -> burned",
			fields: fields{
				protocol: ShelleyProtocol,
				inputs: []TXBuilderInput{
					{
						input: TransactionInput{
							ID:    []byte("input 0"),
							Index: 0,
						},
						amount: 2*ShelleyProtocol.MinimumUtxoValue + 162685,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver.Bytes(),
						Amount:  ShelleyProtocol.MinimumUtxoValue,
					},
				},
			},
		},
		{
			name: "input > output + fee AND change > min utxo AND change - change output fee > min utxo -> sent",
			fields: fields{
				protocol: ShelleyProtocol,
				inputs: []TXBuilderInput{
					{
						input: TransactionInput{
							ID:    []byte("input 0"),
							Index: 0,
						},
						amount: 3 * ShelleyProtocol.MinimumUtxoValue,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver.Bytes(),
						Amount:  ShelleyProtocol.MinimumUtxoValue,
					},
				},
			},
			hasChange: true,
		},
		{
			name: "take in account ttl -> burned",
			fields: fields{
				protocol: ShelleyProtocol,
				inputs: []TXBuilderInput{
					{
						input: TransactionInput{
							ID:    []byte("input 0"),
							Index: 0,
						},
						amount: 2*ShelleyProtocol.MinimumUtxoValue + 164137,
					},
				},
				outputs: []TransactionOutput{
					{
						Address: receiver.Bytes(),
						Amount:  ShelleyProtocol.MinimumUtxoValue,
					},
				},
				ttl: LiveTTL(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &TXBuilder{
				protocol: tt.fields.protocol,
				inputs:   tt.fields.inputs,
				outputs:  tt.fields.outputs,
				ttl:      tt.fields.ttl,
			}
			key := crypto.NewExtendedSigningKey([]byte("change address"), "foo")
			change := NewEnterpriseAddress(key.ExtendedVerificationKey(), Testnet)
			if err := builder.AddFee(change); err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("AddFee() error = %v, wantErr %v", err, tt.wantErr)
			}
			var totalIn uint64
			for _, input := range builder.inputs {
				totalIn += input.amount
			}
			var totalOut uint64
			for _, output := range builder.outputs {
				totalOut += output.Amount
			}
			if got, want := builder.fee+totalOut, totalIn; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			expectedReceiver := receiver
			if tt.hasChange {
				expectedReceiver = change
				if got, want := builder.outputs[0].Amount, builder.protocol.MinimumUtxoValue; got < want {
					t.Errorf("got %v want greater than %v", got, want)
				}
			}
			_, firstOutputReceiver, _ := DecodeAddress(builder.outputs[0].Address)
			if got, want := firstOutputReceiver, expectedReceiver; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		})
	}
}
