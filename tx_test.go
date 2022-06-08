package cardano

import (
	"encoding/hex"
	"math/big"
	"reflect"
	"testing"

	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/internal/cbor"
)

func TestTxEncoding(t *testing.T) {
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
	txBuilder.AddAuxiliaryData(&AuxiliaryData{
		Metadata: Metadata{
			0: map[interface{}]interface{}{
				"secret": "1234",
				"values": uint64(10),
			},
		},
	})

	gotTx := &Tx{}
	wantTx, err := txBuilder.Build()
	if err != nil {
		t.Fatal(err)
	}

	txBytes, err := wantTx.MarshalCBOR()
	if err != nil {
		t.Fatal(err)
	}
	err = gotTx.UnmarshalCBOR(txBytes)
	if err != nil {
		t.Fatal(err)
	}

	for _, txInput := range wantTx.Body.Inputs {
		txInput.Amount = nil
	}

	if !reflect.DeepEqual(wantTx, gotTx) {
		t.Errorf("invalid tx body encoding:\ngot: %+v\nwant: %+v", gotTx, wantTx)
	}
}

func TestCertificateEncoding(t *testing.T) {
	testcases := []struct {
		name    string
		cborHex string
		output  Certificate
	}{
		{
			name:    "StakeRegistration",
			cborHex: "82008200581cd4ffa2b8832507dd670bccff5ec67901737af9dfb2a277d1cf13302b",
			output: Certificate{
				Type: StakeRegistration,
				StakeCredential: StakeCredential{
					Type: KeyCredential,
					KeyHash: AddrKeyHash{
						0xd4, 0xff, 0xa2, 0xb8, 0x83, 0x25, 0x7, 0xdd, 0x67, 0xb, 0xcc, 0xff, 0x5e, 0xc6,
						0x79, 0x1, 0x73, 0x7a, 0xf9, 0xdf, 0xb2, 0xa2, 0x77, 0xd1, 0xcf, 0x13, 0x30, 0x2b,
					},
				},
			},
		},
		{
			name:    "StakeDeregistration",
			cborHex: "82018200581cd4ffa2b8832507dd670bccff5ec67901737af9dfb2a277d1cf13302b",
			output: Certificate{
				Type: StakeDeregistration,
				StakeCredential: StakeCredential{
					Type: KeyCredential,
					KeyHash: AddrKeyHash{
						0xd4, 0xff, 0xa2, 0xb8, 0x83, 0x25, 0x7, 0xdd, 0x67, 0xb, 0xcc, 0xff, 0x5e, 0xc6,
						0x79, 0x1, 0x73, 0x7a, 0xf9, 0xdf, 0xb2, 0xa2, 0x77, 0xd1, 0xcf, 0x13, 0x30, 0x2b,
					},
				},
			},
		},
		{
			name:    "StakeDelegation",
			cborHex: "83028200581cd4ffa2b8832507dd670bccff5ec67901737af9dfb2a277d1cf13302b581c20df8645abddf09403ba2656cda7da2cd163973a5e439c6e43dcbea9",
			output: Certificate{
				Type: StakeDelegation,
				StakeCredential: StakeCredential{
					Type: KeyCredential,
					KeyHash: AddrKeyHash{
						0xd4, 0xff, 0xa2, 0xb8, 0x83, 0x25, 0x7, 0xdd, 0x67, 0xb, 0xcc, 0xff, 0x5e, 0xc6,
						0x79, 0x1, 0x73, 0x7a, 0xf9, 0xdf, 0xb2, 0xa2, 0x77, 0xd1, 0xcf, 0x13, 0x30, 0x2b,
					},
				},
				PoolKeyHash: PoolKeyHash{
					0x20, 0xdf, 0x86, 0x45, 0xab, 0xdd, 0xf0, 0x94, 0x3, 0xba, 0x26, 0x56, 0xcd, 0xa7,
					0xda, 0x2c, 0xd1, 0x63, 0x97, 0x3a, 0x5e, 0x43, 0x9c, 0x6e, 0x43, 0xdc, 0xbe, 0xa9,
				},
			},
		},
		// {
		// 	name:    "PoolRegistration",
		// 	cborHex: "8903581c20df8645abddf09403ba2656cda7da2cd163973a5e439c6e43dcbea9582020df8645abddf09420df8645abddf09420df8645abddf09420df8645abddf0941a001e8480d81e8218230a583901c02e6b0ecdb6bba825ff1fc1e46533c715d5641dccf18cbe06b673e4d4ffa2b8832507dd670bccff5ec67901737af9dfb2a277d1cf13302b81581cd4ffa2b8832507dd670bccff5ec67901737af9dfb2a277d1cf13302b818400190bb844040404045008080808080808080808080808080808f6",
		// 	output: Certificate{
		// 		Type: PoolRegistration,
		// 		Operator: types.PoolKeyHash{
		// 			0x20, 0xdf, 0x86, 0x45, 0xab, 0xdd, 0xf0, 0x94, 0x3, 0xba, 0x26, 0x56, 0xcd, 0xa7,
		// 			0xda, 0x2c, 0xd1, 0x63, 0x97, 0x3a, 0x5e, 0x43, 0x9c, 0x6e, 0x43, 0xdc, 0xbe, 0xa9,
		// 		},
		// 		VrfKeyHash: types.Hash32{
		// 			0x20, 0xdf, 0x86, 0x45, 0xab, 0xdd, 0xf0, 0x94, 0x20, 0xdf, 0x86, 0x45, 0xab, 0xdd, 0xf0, 0x94,
		// 			0x20, 0xdf, 0x86, 0x45, 0xab, 0xdd, 0xf0, 0x94, 0x20, 0xdf, 0x86, 0x45, 0xab, 0xdd, 0xf0, 0x94,
		// 		},
		// 		Pledge: 2000000,
		// 		Margin: types.RationalNumber{P: 35, Q: 10},
		// 		RewardAccount: types.Address{
		// 			B: []byte{
		// 				0x1, 0xc0, 0x2e, 0x6b, 0xe, 0xcd, 0xb6, 0xbb, 0xa8, 0x25, 0xff, 0x1f,
		// 				0xc1, 0xe4, 0x65, 0x33, 0xc7, 0x15, 0xd5, 0x64, 0x1d, 0xcc, 0xf1, 0x8c,
		// 				0xbe, 0x6, 0xb6, 0x73, 0xe4, 0xd4, 0xff, 0xa2, 0xb8, 0x83, 0x25, 0x7, 0xdd,
		// 				0x67, 0xb, 0xcc, 0xff, 0x5e, 0xc6, 0x79, 0x1, 0x73, 0x7a, 0xf9, 0xdf, 0xb2,
		// 				0xa2, 0x77, 0xd1, 0xcf, 0x13, 0x30, 0x2b,
		// 			},
		// 			Hrp: "addr",
		// 		},
		// 		Owners: []types.AddrKeyHash{
		// 			{
		// 				0xd4, 0xff, 0xa2, 0xb8, 0x83, 0x25, 0x7, 0xdd, 0x67, 0xb, 0xcc, 0xff, 0x5e, 0xc6,
		// 				0x79, 0x1, 0x73, 0x7a, 0xf9, 0xdf, 0xb2, 0xa2, 0x77, 0xd1, 0xcf, 0x13, 0x30, 0x2b,
		// 			},
		// 		},
		// 		Relays: []Relay{
		// 			{
		// 				Type: SingleHostAddr,
		// 				Port: types.NewUint64(3000),
		// 				Ipv4: []byte{4, 4, 4, 4},
		// 				Ipv6: []byte{8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8},
		// 			},
		// 		},
		// 	},
		// },
		{
			name:    "PoolRetirement",
			cborHex: "8304581c20df8645abddf09403ba2656cda7da2cd163973a5e439c6e43dcbea919012c",
			output: Certificate{
				Type: PoolRetirement,
				PoolKeyHash: PoolKeyHash{
					0x20, 0xdf, 0x86, 0x45, 0xab, 0xdd, 0xf0, 0x94, 0x3, 0xba, 0x26, 0x56, 0xcd, 0xa7,
					0xda, 0x2c, 0xd1, 0x63, 0x97, 0x3a, 0x5e, 0x43, 0x9c, 0x6e, 0x43, 0xdc, 0xbe, 0xa9,
				},
				Epoch: 300,
			},
		},
		{
			name:    "GenesisKeyDelegation",
			cborHex: "8405581c20df8645abddf09403ba2656cda7da2cd163973a5e439c6e43dcbea9581c20df8645abddf09403ba2656cda7da2cd163973a5e439c6e43dcbea9582020df8645abddf09420df8645abddf09420df8645abddf09420df8645abddf094",
			output: Certificate{
				Type: GenesisKeyDelegation,
				GenesisHash: Hash28{
					0x20, 0xdf, 0x86, 0x45, 0xab, 0xdd, 0xf0, 0x94, 0x3, 0xba, 0x26, 0x56, 0xcd, 0xa7,
					0xda, 0x2c, 0xd1, 0x63, 0x97, 0x3a, 0x5e, 0x43, 0x9c, 0x6e, 0x43, 0xdc, 0xbe, 0xa9,
				},
				GenesisDelegateHash: Hash28{
					0x20, 0xdf, 0x86, 0x45, 0xab, 0xdd, 0xf0, 0x94, 0x3, 0xba, 0x26, 0x56, 0xcd, 0xa7,
					0xda, 0x2c, 0xd1, 0x63, 0x97, 0x3a, 0x5e, 0x43, 0x9c, 0x6e, 0x43, 0xdc, 0xbe, 0xa9,
				},
				VrfKeyHash: Hash32{
					0x20, 0xdf, 0x86, 0x45, 0xab, 0xdd, 0xf0, 0x94, 0x20, 0xdf, 0x86, 0x45, 0xab, 0xdd, 0xf0, 0x94,
					0x20, 0xdf, 0x86, 0x45, 0xab, 0xdd, 0xf0, 0x94, 0x20, 0xdf, 0x86, 0x45, 0xab, 0xdd, 0xf0, 0x94,
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := hex.DecodeString(tc.cborHex)
			if err != nil {
				t.Fatal(err)
			}

			var cert Certificate
			if err := cbor.Unmarshal(data, &cert); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(cert, tc.output) {
				t.Errorf("got: %+v\nwant: %+v", cert, tc.output)
			}

			rb, err := cbor.Marshal(tc.output)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(rb, data) {
				t.Errorf("got: %v\nwant: %v", rb, data)
			}
		})
	}
}

func TestStakeCredentialEncoding(t *testing.T) {
	testcases := []struct {
		name    string
		cborHex string
		output  StakeCredential
	}{
		{
			name:    "AddrKey",
			cborHex: "8200581cd4ffa2b8832507dd670bccff5ec67901737af9dfb2a277d1cf13302b",
			output: StakeCredential{
				Type: KeyCredential,
				KeyHash: AddrKeyHash{
					0xd4, 0xff, 0xa2, 0xb8, 0x83, 0x25, 0x7, 0xdd, 0x67, 0xb, 0xcc, 0xff, 0x5e, 0xc6,
					0x79, 0x1, 0x73, 0x7a, 0xf9, 0xdf, 0xb2, 0xa2, 0x77, 0xd1, 0xcf, 0x13, 0x30, 0x2b,
				},
			},
		},
		{
			name:    "ScriptHash",
			cborHex: "8201581cd4ffa2b8832507dd670bccff5ec67901737af9dfb2a277d1cf13302b",
			output: StakeCredential{
				Type: ScriptCredential,
				ScriptHash: Hash28{
					0xd4, 0xff, 0xa2, 0xb8, 0x83, 0x25, 0x7, 0xdd, 0x67, 0xb, 0xcc, 0xff, 0x5e, 0xc6,
					0x79, 0x1, 0x73, 0x7a, 0xf9, 0xdf, 0xb2, 0xa2, 0x77, 0xd1, 0xcf, 0x13, 0x30, 0x2b,
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := hex.DecodeString(tc.cborHex)
			if err != nil {
				t.Fatal(err)
			}

			var cred StakeCredential
			if err := cbor.Unmarshal(data, &cred); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(cred, tc.output) {
				t.Errorf("got: %+v\nwant: %+v", cred, tc.output)
			}

			rb, err := cbor.Marshal(tc.output)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(rb, data) {
				t.Errorf("got: %v\nwant: %v", rb, data)
			}
		})
	}
}

func TestRelayEncoding(t *testing.T) {
	testcases := []struct {
		name    string
		cborHex string
		output  Relay
	}{
		{
			name:    "SingleHostAddr",
			cborHex: "8400190bb844040404045008080808080808080808080808080808",
			output: Relay{
				Type: SingleHostAddr,
				Port: NewUint64(3000),
				Ipv4: []byte{4, 4, 4, 4},
				Ipv6: []byte{8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8},
			},
		},
		{
			name:    "SingleHostName",
			cborHex: "8301190bb863646e73",
			output: Relay{
				Type:    SingleHostName,
				Port:    NewUint64(3000),
				DNSName: "dns",
			},
		},
		{
			name:    "MultiHostName",
			cborHex: "820263646e73",
			output: Relay{
				Type:    MultiHostName,
				DNSName: "dns",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := hex.DecodeString(tc.cborHex)
			if err != nil {
				t.Fatal(err)
			}

			var r Relay
			if err := cbor.Unmarshal(data, &r); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(r, tc.output) {
				t.Errorf("got: %+v\nwant: %+v", r, tc.output)
			}

			rb, err := cbor.Marshal(tc.output)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(rb, data) {
				t.Errorf("got: %v\nwant: %v", rb, data)
			}
		})
	}
}
