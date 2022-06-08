package cardano

import (
	"errors"
	"math/big"
	"reflect"
	"testing"

	"github.com/echovl/cardano-go/crypto"
)

func TestAssetsEncoding(t *testing.T) {
	assetAmount := 1e9
	assetName := NewAssetName("cardanogo")
	wantAssets := NewAssets().Set(assetName, BigNum(assetAmount))
	gotAssets := NewAssets()
	bytes, err := wantAssets.MarshalCBOR()
	if err != nil {
		t.Fatal(err)
	}
	err = gotAssets.UnmarshalCBOR(bytes)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(wantAssets, gotAssets) {
		t.Errorf("invalid Assets encoding:\ngot: %v\nwant: %v\n", gotAssets, wantAssets)
	}
}

func TestMultiAssetEncoding(t *testing.T) {
	policyKey := crypto.NewXPrvKeyFromEntropy([]byte("policy"), "")
	policyScript, err := NewScriptPubKey(policyKey.PubKey())
	if err != nil {
		t.Fatal(err)
	}
	policyID, err := NewPolicyID(policyScript)
	if err != nil {
		t.Fatal(err)
	}

	assetAmount := int64(1e9)
	assetName := NewAssetName("cardanogo")
	wantMultiAsset := NewMultiAsset().
		Set(
			policyID,
			NewAssets().
				Set(assetName, BigNum(assetAmount)),
		)
	gotMultiAsset := NewMultiAsset()
	bytes, err := wantMultiAsset.MarshalCBOR()
	if err != nil {
		t.Fatal(err)
	}
	err = gotMultiAsset.UnmarshalCBOR(bytes)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(wantMultiAsset, gotMultiAsset) {
		t.Errorf("invalid MultiAsset encoding:\ngot: %v\nwant: %v\n", gotMultiAsset, wantMultiAsset)
	}
}

func TestMintAssetsEncoding(t *testing.T) {
	assetAmount := int64(1e9)
	assetName := NewAssetName("cardanogo")
	wantMintAssets := NewMintAssets().Set(assetName, big.NewInt(assetAmount))
	gotMintAssets := NewMintAssets()
	bytes, err := wantMintAssets.MarshalCBOR()
	if err != nil {
		t.Fatal(err)
	}
	err = gotMintAssets.UnmarshalCBOR(bytes)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(wantMintAssets, gotMintAssets) {
		t.Errorf("invalid MintAssets encoding:\ngot: %v\nwant: %v\n", gotMintAssets, wantMintAssets)
	}
}

func TestMintEncoding(t *testing.T) {
	policyKey := crypto.NewXPrvKeyFromEntropy([]byte("policy"), "")
	policyScript, err := NewScriptPubKey(policyKey.PubKey())
	if err != nil {
		t.Fatal(err)
	}
	policyID, err := NewPolicyID(policyScript)
	if err != nil {
		t.Fatal(err)
	}

	assetAmount := int64(1e9)
	assetName := NewAssetName("cardanogo")
	wantMint := NewMint().
		Set(
			policyID,
			NewMintAssets().
				Set(assetName, big.NewInt(assetAmount)),
		)
	gotMint := NewMint()
	bytes, err := wantMint.MarshalCBOR()
	if err != nil {
		t.Fatal(err)
	}
	err = gotMint.UnmarshalCBOR(bytes)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(wantMint, gotMint) {
		t.Errorf("invalid Mint encoding:\ngot: %v\nwant: %v\n", gotMint, wantMint)
	}
}

func TestValueCmp(t *testing.T) {
	policy1 := NewPolicyIDFromHash([]byte("1234"))
	policy2 := NewPolicyIDFromHash([]byte("4321"))
	token1 := NewAssetName("token1")
	token2 := NewAssetName("token2")
	testcases := []struct {
		name    string
		x       *Value
		y       *Value
		res     int
		wantErr bool
	}{
		{"coin eq", NewValue(10e6), NewValue(10e6), 0, false},
		{"coin lt", NewValue(10e6), NewValue(20e6), -1, false},
		{"coin gt", NewValue(10e6), NewValue(5e6), 1, false},
		{
			"multiAsset lt",
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy1, NewAssets().Set(token1, 5e6))),
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy1, NewAssets().Set(token1, 10e6))),
			-1,
			false,
		},
		{
			"multiAsset eq",
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy2, NewAssets().Set(token2, 10e6))),
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy2, NewAssets().Set(token2, 10e6))),
			0,
			false,
		},
		{
			"multiAsset lt",
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy2, NewAssets().Set(token2, 20e6))),
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy2, NewAssets().Set(token2, 10e6))),
			1,
			false,
		},
		{
			"multiAsset err",
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy1, NewAssets().Set(token2, 20e6))),
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy2, NewAssets().Set(token2, 10e6))),
			1,
			true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.x.Cmp(tc.y)
			if got == 2 {
				if tc.wantErr {
					return
				}
				t.Fatal(errors.New("cannot compare values"))
			}
			if want := tc.res; got != want {
				t.Errorf("invalid cmp\ngot: %v\nwant :%v", got, want)
			}
		})
	}
}

func TestValueArithmetic(t *testing.T) {
	policy1 := NewPolicyIDFromHash([]byte("1234"))
	policy2 := NewPolicyIDFromHash([]byte("4321"))
	token1 := NewAssetName("token1")
	token2 := NewAssetName("token2")

	testcases := []struct {
		name    string
		add     bool
		x       *Value
		y       *Value
		res     *Value
		wantErr bool
	}{
		{"coin add", true, NewValue(10e6), NewValue(10e6), NewValue(20e6), false},
		{"coin add", true, NewValue(10e6), NewValue(10e6), NewValue(20e6), false},
		{"coin sub", false, NewValue(10e6), NewValue(5e6), NewValue(5e6), false},
		{"coin sub underflow", false, NewValue(10e6), NewValue(20e6), NewValue(0), false},
		{
			"multiAsset add same token, same policy",
			true,
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy1, NewAssets().Set(token2, 20e6))),
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy1, NewAssets().Set(token2, 10e6))),
			NewValueWithAssets(20e6, NewMultiAsset().Set(policy1, NewAssets().Set(token2, 30e6))),
			false,
		},
		{
			"multiAsset add diff tokens, same policy",
			true,
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy1, NewAssets().Set(token1, 20e6))),
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy1, NewAssets().Set(token2, 10e6))),
			NewValueWithAssets(
				20e6,
				NewMultiAsset().
					Set(policy1, NewAssets().Set(token1, 20e6).Set(token2, 10e6)),
			),
			false,
		},
		{
			"multiAsset add diff tokens, diff policy",
			true,
			NewValueWithAssets(
				10e6,
				NewMultiAsset().
					Set(policy1, NewAssets().Set(token1, 20e6)),
			),
			NewValueWithAssets(
				10e6,
				NewMultiAsset().
					Set(policy2, NewAssets().Set(token2, 10e6))),
			NewValueWithAssets(
				20e6,
				NewMultiAsset().
					Set(policy1, NewAssets().Set(token1, 20e6)).
					Set(policy2, NewAssets().Set(token2, 10e6)),
			),
			false,
		},
		{
			"multiAsset sub same token, same policy",
			false,
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy1, NewAssets().Set(token2, 20e6))),
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy1, NewAssets().Set(token2, 10e6))),
			NewValueWithAssets(0, NewMultiAsset().Set(policy1, NewAssets().Set(token2, 10e6))),
			false,
		},
		{
			"multiAsset sub diff tokens, same policy",
			false,
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy1, NewAssets().Set(token1, 20e6))),
			NewValueWithAssets(10e6, NewMultiAsset().Set(policy1, NewAssets().Set(token2, 10e6))),
			NewValueWithAssets(
				0,
				NewMultiAsset().
					Set(policy1, NewAssets().Set(token1, 20e6).Set(token2, 0)),
			),
			false,
		},
		{
			"multiAsset sub diff tokens, diff policy",
			false,
			NewValueWithAssets(
				10e6,
				NewMultiAsset().
					Set(policy1, NewAssets().Set(token1, 20e6)),
			),
			NewValueWithAssets(
				10e6,
				NewMultiAsset().
					Set(policy2, NewAssets().Set(token2, 10e6))),
			NewValueWithAssets(
				0,
				NewMultiAsset().
					Set(policy1, NewAssets().Set(token1, 20e6)).
					Set(policy2, NewAssets().Set(token2, 0)),
			),
			false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := &Value{}
			want := tc.res
			if tc.add {
				got = tc.x.Add(tc.y)
			} else {
				got = tc.x.Sub(tc.y)
			}

			equal := got.Cmp(want)
			if equal == 2 {
				t.Fatal(errors.New("cannot compare values"))
			}

			if equal != 0 {
				t.Errorf("invalid Add\ngot: %v\nwant: %v", got, want)
			}
		})
	}
}
