package cardano

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/echovl/cardano-go/internal/cbor"
	"math/big"
	"testing"

	"github.com/echovl/cardano-go/crypto"
)

var alonzoProtocol = &ProtocolParams{
	CoinsPerUTXOWord: 4310,
	MinFeeA:          44,
	MinFeeB:          155381,
}

// See https://github.com/input-output-hk/cardano-ledger/blob/master/doc/explanations/min-utxo-alonzo.rst
func TestMinUTXO(t *testing.T) {
	pubKeys := []crypto.PubKey{
		crypto.NewXPrvKeyFromEntropy([]byte("pol1"), "").PubKey(),
		crypto.NewXPrvKeyFromEntropy([]byte("pol2"), "").PubKey(),
	}

	testcases := []struct {
		name     string
		policies []int
		assets   []string
		minUTXO  Coin
	}{
		{
			name:     "One policyID, one 0-character asset name",
			policies: []int{0},
			assets:   []string{""},
			minUTXO:  Coin(1310316),
		},
		{
			name:     "One policyID, one 1-character asset name",
			policies: []int{0},
			assets:   []string{"a"},
			minUTXO:  Coin(1344798),
		},
		{
			name:     "One policyID, three 1-character asset name",
			policies: []int{0},
			assets:   []string{"a", "b", "c"},
			minUTXO:  Coin(1448244),
		},
		{
			name:     "Two policyIDs, one 0-character asset name",
			policies: []int{0, 1},
			assets:   []string{""},
			minUTXO:  Coin(1482726),
		},
		{
			name:     "Two policyIDs, one 1-character asset name",
			policies: []int{0, 1},
			assets:   []string{"a"},
			minUTXO:  Coin(1517208),
		},
		// ## Impossible to have 96 1-character asset name
		// {
		// 	name:     "Three policyIDs, ninety-six 1-character names between them (total)",
		// 	policies: []int{0, 1, 2},
		// 	assets: []string{
		// 		"a", "b", "c", ...,
		// 	},
		// 	minUTXO: Coin(6896400),
		// },
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
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
		})
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

func TestMinCoins(t *testing.T)  {
	txBuilder := NewTxBuilder(alonzoProtocol)
	addr, _ := NewAddress("DdzFFzCqrht7FAf8MpryP1p8sgkmFRUnDpifnnu4ZxpBjbCTSDwJVAaDsDqrC7SLYFx8fUrDcNNsD4AMiUgg2wGywTVpcfB1F3AHrGkv")


	name, _ := hex.DecodeString("4d494e")
	policyHash, _ := NewHash28("29d222ce763455e3d7a09a665ce554f00ac89d2e99a1a83d267170c6")
	assetName := NewAssetName(string(name))
	policyId := NewPolicyIDFromHash(policyHash)

	asset := NewMultiAsset().Set(policyId,  NewAssets().Set(assetName, BigNum(5000000)))
	name, _ = hex.DecodeString("69555344")
	policyHash, _ = NewHash28("f66d78b4a3cb3d37afa0ec36461e51ecbde00f26c8f0a68f94b69880")
	assetName = NewAssetName(string(name))
	policyId = NewPolicyIDFromHash(policyHash)
	asset.Set(policyId,  NewAssets().Set(assetName, BigNum(100000)))

	//asset = nil
	_ = asset
	output := NewTxOutput(addr,
		NewValue(Coin(1000000)),
		//NewValueWithAssets(Coin(1200000), asset),
		)
	fmt.Println(txBuilder.MinCoinsForTxOut(output))
}

func TestTxOut(t *testing.T)  {
	testData := []struct{
		address string
		amount uint64
		want string
	}{
		{
			address: "addr_test1qqth544yyqh8ahg0899ms59emls89cs9l9ra0n9nlrwtgahppgsq2ykgpqpgewlkwkyhsqn29k8dxp7xthncvfwht9kqdaq2fy",
			amount:  100,
			want:    "82583900177a56a4202e7edd0f394bb850b9dfe072e205f947d7ccb3f8dcb476e10a200512c808028cbbf6758978026a2d8ed307c65de78625d7596c1864",
		},
		{
			address: "Ae2tdPwUPEZEJsmHfRC3gutyFZuJSXLumCu1oovMSps3EVewbqoHDnoZm3d",
			amount:  100,
			want:    "82582b82d818582183581caf67c9c546e2a69870cc7c2bc5f6633253dfa36143ffddfda4523ee6a0001a78f496c81864",
		},
		{
			address: "DdzFFzCqrht7FAf8MpryP1p8sgkmFRUnDpifnnu4ZxpBjbCTSDwJVAaDsDqrC7SLYFx8fUrDcNNsD4AMiUgg2wGywTVpcfB1F3AHrGkv",
			amount:  100,
			want:    "82584c82d818584283581cd1ed7fa89a505f9c83de0ef1178b383258e71b2375b4e1279d5e0e0fa101581e581cc655d01a842282dd46918c9198d9e98efad24af87b9b0205aed69fd4001a237a87d71864",
		},
	}

	for _, data := range testData {
		addr, _ := NewAddress(data.address)
		out := NewTxOutput(
			addr,
			NewValue(Coin(data.amount)),
			//cardano.NewValueWithAssets(cardano.Coin(sendAmount), newAsset.MultiAsset()),
		)

		rawbytes, err := cbor.Marshal(out)
		out1 := new(TxOutput)
		fmt.Println(out1.UnmarshalCBOR(rawbytes))
		fmt.Println(out1.Amount, out1.Address)

		fmt.Println(err, hex.EncodeToString(rawbytes))

		fmt.Println(hex.EncodeToString(rawbytes) == data.want)
	}
}

func TestTxOutWithAsset(t *testing.T)  {
	var testdata = []struct {
		address     string
		amount      uint64
		assetId     string
		assetAmount uint64
		want        string
	}{
		{
			address:     "addr_test1vqpjd93t42ju4majh9tcz69z2fvmaeyxzxvpr3x95g9mw4sxmvk7w",
			amount:      100,
			assetId:     "77e7a4688886467574045c6a0126d140644b97e12b644d473e71d3e9315354",
			assetAmount: 10,
			want:        "82581d600326962baaa5caefb2b9578168a25259bee486119811c4c5a20bb756821864a1581c77e7a4688886467574045c6a0126d140644b97e12b644d473e71d3e9a1433153540a",
		},
		{
			address:     "addr_test1vqpjd93t42ju4majh9tcz69z2fvmaeyxzxvpr3x95g9mw4sxmvk7w",
			amount:      100,
			assetId:     "77e7a4688886467574045c6a0126d140644b97e12b644d473e71d3e9",
			assetAmount: 10,
			want:        "82581d600326962baaa5caefb2b9578168a25259bee486119811c4c5a20bb756821864a1581c77e7a4688886467574045c6a0126d140644b97e12b644d473e71d3e9a1400a",
		},
		{
			address:     "addr1q860w3lz6zd9tn0uqcvw782qmcs6lnztz2asy0555sr2kvh57ar795y62hxlcpscauw5ph3p4lxyky4mqglfffqx4veq6gs3hg",
			amount:      100,
			assetId:     "77e7a4688886467574045c6a0126d140644b97e12b644d473e71d3e9315354",
			assetAmount: 10,
			want:        "82583901f4f747e2d09a55cdfc0618ef1d40de21afcc4b12bb023e94a406ab32f4f747e2d09a55cdfc0618ef1d40de21afcc4b12bb023e94a406ab32821864a1581c77e7a4688886467574045c6a0126d140644b97e12b644d473e71d3e9a1433153540a",
		},
		{
			address:     "Ae2tdPwUPEZEJsmHfRC3gutyFZuJSXLumCu1oovMSps3EVewbqoHDnoZm3d",
			amount:      100,
			assetId:     "77e7a4688886467574045c6a0126d140644b97e12b644d473e71d3e9315354",
			assetAmount: 10,
			want:        "82582b82d818582183581caf67c9c546e2a69870cc7c2bc5f6633253dfa36143ffddfda4523ee6a0001a78f496c8821864a1581c77e7a4688886467574045c6a0126d140644b97e12b644d473e71d3e9a1433153540a",
		},
		{
			address:     "DdzFFzCqrht7FAf8MpryP1p8sgkmFRUnDpifnnu4ZxpBjbCTSDwJVAaDsDqrC7SLYFx8fUrDcNNsD4AMiUgg2wGywTVpcfB1F3AHrGkv",
			amount:      100,
			assetId:     "77e7a4688886467574045c6a0126d140644b97e12b644d473e71d3e9315354",
			assetAmount: 10,
			want:        "82584c82d818584283581cd1ed7fa89a505f9c83de0ef1178b383258e71b2375b4e1279d5e0e0fa101581e581cc655d01a842282dd46918c9198d9e98efad24af87b9b0205aed69fd4001a237a87d7821864a1581c77e7a4688886467574045c6a0126d140644b97e12b644d473e71d3e9a1433153540a",
		},
	}

	for _, data := range testdata {
		addr, err := NewAddress(data.address)
		asset, err := ParseAssetId(data.assetId, data.assetAmount)
		out := NewTxOutput(
			addr,
			NewValueWithAssets(Coin(data.amount), asset),
		)

		rawbytes, err := cbor.Marshal(out)
		out1 := new(TxOutput)
		fmt.Println(out1.UnmarshalCBOR(rawbytes))
		fmt.Println(out1.Amount, out1.Address, out1.Address.String() == addr.String())

		fmt.Println(err, hex.EncodeToString(rawbytes))

		fmt.Println(hex.EncodeToString(rawbytes) == data.want)
	}
}

func ParseAssetId(assetId string, assetNum uint64) (*MultiAsset, error) {
	bytes, err := hex.DecodeString(assetId)
	if err != nil {
		return nil, err
	}

	if len(bytes) < 28 || len(bytes) > 60 {
		return nil, errors.New("invalid asset id length")
	}

	policyHash := make([]byte, 28)
	copy(policyHash, bytes[:28])
	assetName := NewAssetName(string(bytes[28:]))
	policyId := NewPolicyIDFromHash(policyHash)
	//fmt.Println(policyHash, policyId, assetName)
	return NewMultiAsset().Set(policyId, NewAssets().Set(assetName, BigNum(assetNum))), nil
}
