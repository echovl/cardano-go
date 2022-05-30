package tx

import (
	"fmt"
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
	key := crypto.NewXPrvKeyFromEntropy([]byte("receiver address"), "foo")
	payment, err := types.NewAddrKeyCredential(key.PubKey())
	if err != nil {
		t.Fatal(err)
	}
	receiver, err := types.NewEnterpriseAddress(types.Testnet, payment)
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		tx       Tx
		protocol *ProtocolParams
		inputs   []*TxInput
		outputs  []*TxOutput
		ttl      uint64
		fee      uint64
		vkeys    map[string]crypto.XPubKey
		pkeys    map[string]crypto.XPrvKey
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
				inputs: []*TxInput{
					{
						TxHash: make([]byte, 32),
						Index:  uint64(0),
						Amount: 200000,
					},
				},
				outputs: []*TxOutput{
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
				inputs: []*TxInput{
					{
						TxHash: make([]byte, 32),
						Index:  0,
						Amount: minUTXO(nil, shelleyProtocol) + 165501,
					},
				},
				outputs: []*TxOutput{
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
				inputs: []*TxInput{
					{
						TxHash: make([]byte, 32),
						Index:  0,
						Amount: 2*minUTXO(nil, shelleyProtocol) - 1,
					},
				},
				outputs: []*TxOutput{
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
				inputs: []*TxInput{
					{
						TxHash: make([]byte, 32),
						Index:  0,
						Amount: 2*minUTXO(nil, shelleyProtocol) + 162685,
					},
				},
				outputs: []*TxOutput{
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
				inputs: []*TxInput{
					{
						TxHash: make([]byte, 32),
						Index:  0,
						Amount: 3 * minUTXO(nil, shelleyProtocol),
					},
				},
				outputs: []*TxOutput{
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
				inputs: []*TxInput{
					{
						TxHash: make([]byte, 32),
						Index:  0,
						Amount: 2*minUTXO(nil, shelleyProtocol) + 164137,
					},
				},
				outputs: []*TxOutput{
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
			key := crypto.NewXPrvKeyFromEntropy([]byte("change address"), "foo")
			payment, err := types.NewAddrKeyCredential(key.PubKey())
			if err != nil {
				t.Fatal(err)
			}
			change, err := types.NewEnterpriseAddress(types.Testnet, payment)
			if err != nil {
				t.Fatal(err)
			}
			builder := NewTxBuilder(shelleyProtocol)
			builder.AddInputs(tt.fields.inputs...)
			builder.AddOutputs(tt.fields.outputs...)
			builder.SetTTL(tt.fields.ttl)
			builder.Sign(key.Bech32("addr_xsk"))
			if err := builder.AddChangeIfNeeded(change.Bech32()); err != nil {
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
			if got, want := firstOutputReceiver.Bech32(), expectedReceiver.Bech32(); got != want {
				for _, out := range builder.tx.Body.Outputs {
					fmt.Printf("%+v", out)
				}
				t.Errorf("got %v want %v", got, want)
			}
		})
	}
}
