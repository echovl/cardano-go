package cardano

import (
	"math/big"
	"testing"

	"github.com/echovl/cardano-go/crypto"
)

var alonzoProtocol = &ProtocolParams{
	CoinsPerUTXOWord: 34482,
	MinFeeA:          44,
	MinFeeB:          155381,
}

func TestMinUTXO(t *testing.T) {
	pubKeys := []crypto.PubKey{
		crypto.NewXPrvKeyFromEntropy([]byte("pol1"), "").PubKey(),
		crypto.NewXPrvKeyFromEntropy([]byte("pol2"), "").PubKey(),
	}

	testcases := []struct {
		policies []int
		assets   []string
		minUTXO  Coin
	}{
		{
			policies: []int{0},
			assets:   []string{""},
			minUTXO:  Coin(utxoEntrySizeWithoutVal+11) * alonzoProtocol.CoinsPerUTXOWord,
		},
		{
			policies: []int{0},
			assets:   []string{"a"},
			minUTXO:  Coin(utxoEntrySizeWithoutVal+12) * alonzoProtocol.CoinsPerUTXOWord,
		},
		{
			policies: []int{0},
			assets:   []string{"a", "b", "c"},
			minUTXO:  Coin(utxoEntrySizeWithoutVal+15) * alonzoProtocol.CoinsPerUTXOWord,
		},
		{
			policies: []int{0, 1},
			assets:   []string{"a"},
			minUTXO:  Coin(utxoEntrySizeWithoutVal+17) * alonzoProtocol.CoinsPerUTXOWord,
		},
	}

	for _, tc := range testcases {
		multiAsset := NewMultiAsset()
		for _, policy := range tc.policies {
			script, err := NewScriptPubKey(pubKeys[policy])
			if err != nil {
				t.Fatal(err)
			}
			assets := NewAssets()
			policyID, err := NewPolicyID(script)
			if err != nil {
				t.Fatal(err)
			}
			for _, assetName := range tc.assets {
				assets.Set(NewAssetName(assetName), 0)
			}
			multiAsset.Set(policyID, assets)
		}

		txBuilder := NewTxBuilder(alonzoProtocol)

		txOutput := &TxOutput{Amount: NewValueWithAssets(0, multiAsset)}
		got := txBuilder.MinCoinsForTxOut(txOutput)
		want := tc.minUTXO

		if got != want {
			t.Errorf("invalid minUTXO\ngot: %d\nwant: %d", got, want)
		}
	}
}

func TestSimpleTx(t *testing.T) {
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
			txBuilder.Sign(sk)

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

func TestMintingAssets(t *testing.T) {
	txBuilder := NewTxBuilder(alonzoProtocol)

	paymentKey := crypto.NewXPrvKeyFromEntropy([]byte("payment"), "")
	policyKey := crypto.NewXPrvKeyFromEntropy([]byte("policy"), "")

	txHash, err := NewHash32("030858db80bf94041b7b1c6fbc0754a9bd7113ec9025b1157a9a4e02135f3518")
	if err != nil {
		t.Fatal(err)
	}
	addr, err := NewAddress("addr_test1vp9uhllavnhwc6m6422szvrtq3eerhleer4eyu00rmx8u6c42z3v8")
	if err != nil {
		t.Fatal(err)
	}

	policyScript, err := NewScriptPubKey(policyKey.PubKey())
	if err != nil {
		t.Fatal(err)
	}
	policyID, err := NewPolicyID(policyScript)
	if err != nil {
		t.Fatal(err)
	}

	inputAmount, transferAmount, assetAmount := Coin(1e9), Coin(10e6), int64(1e9)

	assetName := NewAssetName("cardanogo")
	newAsset := NewMint().
		Set(
			policyID,
			NewMintAssets().
				Set(assetName, big.NewInt(assetAmount)),
		)

	txBuilder.AddInputs(
		NewTxInput(txHash, 0, NewValue(inputAmount)),
	)
	txBuilder.AddOutputs(
		NewTxOutput(addr, NewValueWithAssets(transferAmount, newAsset.MultiAsset())),
	)

	txBuilder.Mint(newAsset)
	txBuilder.AddNativeScript(policyScript)
	txBuilder.SetTTL(100000)
	txBuilder.Sign(paymentKey.PrvKey())
	txBuilder.Sign(policyKey.PrvKey())
	txBuilder.AddChangeIfNeeded(addr)
	tx, err := txBuilder.Build()
	if err != nil {
		t.Fatal(err)
	}

	minFee, err := txBuilder.MinFee()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := tx.Body.Fee, minFee; got != want {
		t.Errorf("invalid tx fee:\ngot: %v\nwant: %v", got, want)
	}

	wantAmount := NewValueWithAssets(transferAmount, newAsset.MultiAsset())
	gotAmount := tx.Body.Outputs[1].Amount
	if gotAmount.Cmp(wantAmount) != 0 {
		t.Errorf("invalid output asset amount:\ngot: %v\nwant: %v", gotAmount, wantAmount)
	}

}

func TestSendingMultiAssets(t *testing.T) {
	paymentKey := crypto.NewXPrvKeyFromEntropy([]byte("payment"), "")
	policyKey := crypto.NewXPrvKeyFromEntropy([]byte("policy"), "")
	policyScript, err := NewScriptPubKey(policyKey.PubKey())
	if err != nil {
		t.Fatal(err)
	}
	policyID, err := NewPolicyID(policyScript)
	if err != nil {
		t.Fatal(err)
	}
	assetName := NewAssetName("cardanogo")
	txHash, err := NewHash32("030858db80bf94041b7b1c6fbc0754a9bd7113ec9025b1157a9a4e02135f3518")
	if err != nil {
		t.Fatal(err)
	}
	addr, err := NewAddress("addr_test1vp9uhllavnhwc6m6422szvrtq3eerhleer4eyu00rmx8u6c42z3v8")
	if err != nil {
		t.Fatal(err)
	}

	testcases := []struct {
		name           string
		txInputAmount  *Value
		txOutputAmount *Value
		wantErr        bool
	}{
		{
			"partial asset transfer",
			NewValueWithAssets(
				10e6,
				NewMultiAsset().Set(policyID, NewAssets().Set(assetName, 10e6)),
			),
			NewValueWithAssets(
				5e6,
				NewMultiAsset().Set(policyID, NewAssets().Set(assetName, 5e6)),
			),
			false,
		},
		{
			"full asset transfer",
			NewValueWithAssets(
				10e6,
				NewMultiAsset().Set(policyID, NewAssets().Set(assetName, 10e6)),
			),
			NewValueWithAssets(
				5e6,
				NewMultiAsset().Set(policyID, NewAssets().Set(assetName, 10e6)),
			),
			false,
		},
		{
			"only ada transfer",
			NewValueWithAssets(
				10e6,
				NewMultiAsset().Set(policyID, NewAssets().Set(assetName, 10e6)),
			),
			NewValueWithAssets(
				5e6,
				NewMultiAsset().Set(policyID, NewAssets().Set(assetName, 0)),
			),
			false,
		},
		{
			"output asset > input asset",
			NewValueWithAssets(
				10e6,
				NewMultiAsset().Set(policyID, NewAssets().Set(assetName, 10e6)),
			),
			NewValueWithAssets(
				5e6,
				NewMultiAsset().Set(policyID, NewAssets().Set(assetName, 20e6)),
			),
			true,
		},
		{
			"insuficient change coins",
			NewValueWithAssets(
				10e6,
				NewMultiAsset().Set(policyID, NewAssets().Set(assetName, 10e6)),
			),
			NewValueWithAssets(
				98e5,
				NewMultiAsset().Set(policyID, NewAssets().Set(assetName, 5e6)),
			),
			true,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			txBuilder := NewTxBuilder(alonzoProtocol)

			txBuilder.AddInputs(
				NewTxInput(txHash, 0, tc.txInputAmount),
			)
			txBuilder.AddOutputs(
				NewTxOutput(addr, tc.txOutputAmount),
			)
			txBuilder.SetTTL(100000)
			txBuilder.Sign(paymentKey.PrvKey())
			txBuilder.Sign(policyKey.PrvKey())
			txBuilder.AddAuxiliaryData(&AuxiliaryData{
				Metadata: Metadata{
					0: map[string]interface{}{
						"hello": "cardano-go",
					},
				},
			})

			txBuilder.AddChangeIfNeeded(addr)
			tx, err := txBuilder.Build()
			if err != nil {
				if tc.wantErr {
					return
				}
				t.Fatal(err)
			}

			gotAmount := tx.Body.Outputs[1].Amount
			if gotAmount.Cmp(tc.txOutputAmount) != 0 {
				t.Errorf("invalid output asset amount:\ngot: %v\nwant: %v", gotAmount, tc.txOutputAmount)
			}

			minFee, err := txBuilder.MinFee()
			if err != nil {
				t.Fatal(err)
			}

			if got, want := tx.Body.Fee, minFee; got != want {
				t.Errorf("invalid tx fee:\ngot: %v\nwant: %v", got, want)
			}

		})
	}
}

func TestAddChangeIfNeeded(t *testing.T) {
	txBuilder := NewTxBuilder(alonzoProtocol)
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
						Amount: NewValue(txBuilder.MinCoinsForTxOut(emptyTxOut) + 165501),
					},
				},
				outputs: []*TxOutput{
					{
						Address: receiver,
						Amount:  NewValue(txBuilder.MinCoinsForTxOut(emptyTxOut)),
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
						Amount: NewValue(2*txBuilder.MinCoinsForTxOut(emptyTxOut) - 1),
					},
				},
				outputs: []*TxOutput{
					{
						Address: receiver,
						Amount:  NewValue(txBuilder.MinCoinsForTxOut(emptyTxOut)),
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
						Amount: NewValue(2*txBuilder.MinCoinsForTxOut(emptyTxOut) + 162685),
					},
				},
				outputs: []*TxOutput{
					{
						Address: receiver,
						Amount:  NewValue(txBuilder.MinCoinsForTxOut(emptyTxOut)),
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
						Amount: NewValue(3 * txBuilder.MinCoinsForTxOut(emptyTxOut)),
					},
				},
				outputs: []*TxOutput{
					{
						Address: receiver,
						Amount:  NewValue(txBuilder.MinCoinsForTxOut(emptyTxOut)),
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
						Amount: NewValue(2*txBuilder.MinCoinsForTxOut(emptyTxOut) + 164137),
					},
				},
				outputs: []*TxOutput{
					{
						Address: receiver,
						Amount:  NewValue(txBuilder.MinCoinsForTxOut(emptyTxOut)),
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
			changeAddr, err := NewEnterpriseAddress(Testnet, payment)
			if err != nil {
				t.Fatal(err)
			}
			txBuilder := NewTxBuilder(alonzoProtocol)
			txBuilder.AddInputs(tc.fields.inputs...)
			txBuilder.AddOutputs(tc.fields.outputs...)
			txBuilder.SetTTL(tc.fields.ttl)
			txBuilder.Sign(key.PrvKey())
			txBuilder.AddChangeIfNeeded(changeAddr)
			tx, err := txBuilder.Build()
			if err != nil {
				if tc.wantErr {
					return
				}
				t.Fatal(err)
			}
			var totalIn Coin
			for _, input := range tx.Body.Inputs {
				totalIn += input.Amount.Coin
			}
			var totalOut Coin
			for _, output := range tx.Body.Outputs {
				totalOut += output.Amount.Coin
			}
			if got, want := tx.Body.Fee+totalOut, totalIn; got != want {
				t.Errorf("invalid fee+totalOut: got %v want %v", got, want)
			}
			expectedReceiver := receiver
			if tc.hasChange {
				expectedReceiver = changeAddr
				if got, want := tx.Body.Outputs[0].Amount, txBuilder.MinCoinsForTxOut(emptyTxOut); got.Coin < want {
					t.Errorf("invalid change output: got %v want greater than %v", got, want)
				}
			}
			firstOutputReceiver := tx.Body.Outputs[0].Address
			if got, want := firstOutputReceiver.Bech32(), expectedReceiver.Bech32(); got != want {
				t.Errorf("invalid change output receiver: got %v want %v", got, want)
			}
		})
	}
}
