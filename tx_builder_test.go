package cardano

import (
	"testing"

	"github.com/echovl/cardano-go/crypto"
)

var alonzoProtocol = &ProtocolParams{
	CoinsPerUTXOWord: 34482,
	MinFeeA:          44,
	MinFeeB:          155381,
}

func TestSimpleBuild(t *testing.T) {
	testcases := []struct {
		name     string
		sk       string
		txHashIn string
		addrOut  string
		input    Coin
		output   Coin
		fee      uint64
		txHash   string
		wantErr  bool
	}{
		{
			name:     "ok",
			sk:       "addr_sk1uqpfmhkflccgy9wzdrgshtjez963a0rj2apjxzga9dysw5y4tap0ame6lckwe94wq68dyc2669vp7e64rhmd0lmyf0gy3k7aeqt5dcc8x0qzj",
			txHashIn: "7a040587157289e80e524710021fa9a61d22a597b70786f21a4a78b61dddee29",
			addrOut:  "addr1qxn0t7jnv8lrdd5xa6mlcap6qf8ln08pc6k8qxa7un0new2pkrthnm4f5hn6eg3nju6jn6l3994ucy099cw42xu7rmjq8l960u",
			input:    100 * 1e6,
			output:   99 * 1e6,
			fee:      1 * 1e6,
			txHash:   "b59fec079542f4785d3d197ada365e496de932237ae168cba599926dd6f42e31",
			wantErr:  false,
		},
		{
			name:     "insuficient output",
			sk:       "addr_sk1uqpfmhkflccgy9wzdrgshtjez963a0rj2apjxzga9dysw5y4tap0ame6lckwe94wq68dyc2669vp7e64rhmd0lmyf0gy3k7aeqt5dcc8x0qzj",
			txHashIn: "7a040587157289e80e524710021fa9a61d22a597b70786f21a4a78b61dddee29",
			addrOut:  "addr1qxn0t7jnv8lrdd5xa6mlcap6qf8ln08pc6k8qxa7un0new2pkrthnm4f5hn6eg3nju6jn6l3994ucy099cw42xu7rmjq8l960u",
			input:    101 * 1e6,
			output:   99 * 1e6,
			fee:      1 * 1e6,
			wantErr:  true,
		},
		{
			name:     "insuficient input",
			sk:       "addr_sk1uqpfmhkflccgy9wzdrgshtjez963a0rj2apjxzga9dysw5y4tap0ame6lckwe94wq68dyc2669vp7e64rhmd0lmyf0gy3k7aeqt5dcc8x0qzj",
			txHashIn: "7a040587157289e80e524710021fa9a61d22a597b70786f21a4a78b61dddee29",
			addrOut:  "addr1qxn0t7jnv8lrdd5xa6mlcap6qf8ln08pc6k8qxa7un0new2pkrthnm4f5hn6eg3nju6jn6l3994ucy099cw42xu7rmjq8l960u",
			input:    99 * 1e6,
			output:   99 * 1e6,
			fee:      1 * 1e6,
			wantErr:  true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			addrOut, err := NewAddress(tc.addrOut)
			if err != nil {
				t.Fatal(err)
			}
			sk, err := crypto.NewPrvKey(tc.sk)
			if err != nil {
				t.Fatal(err)
			}
			txHashIn, err := NewHash32(tc.txHashIn)
			if err != nil {
				t.Fatal(err)
			}
			txIn := NewTxInput(txHashIn, 0, NewValue(tc.input))
			txOut := NewTxOutput(addrOut, NewValue(tc.output))

			txBuilder := NewTxBuilder(alonzoProtocol)

			txBuilder.AddInputs(txIn)
			txBuilder.AddOutputs(txOut)
			txBuilder.SetFee(Coin(tc.fee))

			if err := txBuilder.Sign(sk); err != nil {
				t.Fatal(err)
			}

			tx, err := txBuilder.Build()
			if err != nil {
				if tc.wantErr {
					return
				}
				t.Fatal(err)
			}

			txHash, err := tx.Hash()
			if err != nil {
				t.Fatal(err)
			}

			if got, want := txHash.String(), tc.txHash; got != want {
				t.Errorf("wrong tx hash\ngot: %s\nwant: %s", got, want)
			}

		})
	}

}

func TestAddChangeIfNeeded(t *testing.T) {
	key := crypto.NewXPrvKeyFromEntropy([]byte("receiver address"), "foo")
	payment, err := NewKeyCredential(key.PubKey())
	if err != nil {
		t.Fatal(err)
	}
	receiver, err := NewEnterpriseAddress(Testnet, payment)
	if err != nil {
		t.Fatal(err)
	}

	emptyTxOut := &TxOutput{Amount: NewValue(0)}

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

	testcases := []struct {
		name      string
		fields    fields
		hasChange bool
		wantErr   bool
	}{
		{
			name: "input < output + fee",
			fields: fields{
				protocol: alonzoProtocol,
				inputs: []*TxInput{
					{
						TxHash: make([]byte, 32),
						Index:  uint64(0),
						Amount: NewValue(200000),
					},
				},
				outputs: []*TxOutput{
					{
						Address: receiver,
						Amount:  NewValue(200000),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "input == output + fee",
			fields: fields{
				protocol: alonzoProtocol,
				inputs: []*TxInput{
					{
						TxHash: make([]byte, 32),
						Index:  0,
						Amount: NewValue(minUTXO(emptyTxOut, alonzoProtocol) + 165501),
					},
				},
				outputs: []*TxOutput{
					{
						Address: receiver,
						Amount:  NewValue(minUTXO(emptyTxOut, alonzoProtocol)),
					},
				},
			},
		},
		{
			name: "input > output + fee AND change < min utxo value -> burned",
			fields: fields{
				protocol: alonzoProtocol,
				inputs: []*TxInput{
					{
						TxHash: make([]byte, 32),
						Index:  0,
						Amount: NewValue(2*minUTXO(emptyTxOut, alonzoProtocol) - 1),
					},
				},
				outputs: []*TxOutput{
					{
						Address: receiver,
						Amount:  NewValue(minUTXO(emptyTxOut, alonzoProtocol)),
					},
				},
			},
		},
		{
			name: "input > output + fee AND change > min utxo BUT change - change output fee < min utxo -> burned",
			fields: fields{
				protocol: alonzoProtocol,
				inputs: []*TxInput{
					{
						TxHash: make([]byte, 32),
						Index:  0,
						Amount: NewValue(2*minUTXO(emptyTxOut, alonzoProtocol) + 162685),
					},
				},
				outputs: []*TxOutput{
					{
						Address: receiver,
						Amount:  NewValue(minUTXO(emptyTxOut, alonzoProtocol)),
					},
				},
			},
		},
		{
			name: "input > output + fee AND change > min utxo AND change - change output fee > min utxo -> sent",
			fields: fields{
				protocol: alonzoProtocol,
				inputs: []*TxInput{
					{
						TxHash: make([]byte, 32),
						Index:  0,
						Amount: NewValue(3 * minUTXO(emptyTxOut, alonzoProtocol)),
					},
				},
				outputs: []*TxOutput{
					{
						Address: receiver,
						Amount:  NewValue(minUTXO(emptyTxOut, alonzoProtocol)),
					},
				},
			},
			hasChange: true,
		},
		{
			name: "take in account ttl -> burned",
			fields: fields{
				protocol: alonzoProtocol,
				inputs: []*TxInput{
					{
						TxHash: make([]byte, 32),
						Index:  0,
						Amount: NewValue(2*minUTXO(emptyTxOut, alonzoProtocol) + 164137),
					},
				},
				outputs: []*TxOutput{
					{
						Address: receiver,
						Amount:  NewValue(minUTXO(emptyTxOut, alonzoProtocol)),
					},
				},
				ttl: 100,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			key := crypto.NewXPrvKeyFromEntropy([]byte("change address"), "foo")
			payment, err := NewKeyCredential(key.PubKey())
			if err != nil {
				t.Fatal(err)
			}
			change, err := NewEnterpriseAddress(Testnet, payment)
			if err != nil {
				t.Fatal(err)
			}
			builder := NewTxBuilder(alonzoProtocol)
			builder.AddInputs(tc.fields.inputs...)
			builder.AddOutputs(tc.fields.outputs...)
			builder.SetTTL(tc.fields.ttl)
			builder.Sign(key.PrvKey())
			if err := builder.AddChangeIfNeeded(change); err != nil {
				if tc.wantErr {
					return
				}
				t.Fatalf("AddFee() error = %v, wantErr %v", err, tc.wantErr)
			}
			var totalIn Coin
			for _, input := range builder.tx.Body.Inputs {
				totalIn += input.Amount.Coin
			}
			var totalOut Coin
			for _, output := range builder.tx.Body.Outputs {
				totalOut += output.Amount.Coin
			}
			if got, want := builder.tx.Body.Fee+totalOut, totalIn; got != want {
				t.Errorf("invalid fee+totalOut: got %v want %v", got, want)
			}
			expectedReceiver := receiver
			if tc.hasChange {
				expectedReceiver = change
				if got, want := builder.tx.Body.Outputs[0].Amount, minUTXO(emptyTxOut, builder.protocol); got.Coin < want {
					t.Errorf("invalid change output: got %v want greater than %v", got, want)
				}
			}
			firstOutputReceiver := builder.tx.Body.Outputs[0].Address
			if got, want := firstOutputReceiver.Bech32(), expectedReceiver.Bech32(); got != want {
				t.Errorf("invalid change output receiver: got %v want %v", got, want)
			}
		})
	}
}
