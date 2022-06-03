package cardano

import (
	"testing"

	"github.com/echovl/cardano-go/crypto"
)

func TestValueCmp(t *testing.T) {
	testcases := []struct {
		x   *Value
		y   *Value
		res int
	}{
		{
			x:   NewValue(10e6),
			y:   NewValue(10e6),
			res: 0,
		},
		{
			x:   NewValue(10e6),
			y:   NewValue(20e6),
			res: -1,
		},
		{
			x:   NewValue(10e6),
			y:   NewValue(5e6),
			res: 1,
		},
	}

	for _, tc := range testcases {
		got, err := tc.x.Cmp(tc.y)
		if err != nil {
			t.Fatal(err)
		}
		if want := tc.res; got != want {
			t.Errorf("invalid cmp\ngot: %v\nwant :%v", got, want)
		}
	}
}

func TestValueArithmetic(t *testing.T) {
	pubKeys := []crypto.PubKey{
		crypto.NewXPrvKeyFromEntropy([]byte("pol1"), "").PubKey(),
		crypto.NewXPrvKeyFromEntropy([]byte("pol2"), "").PubKey(),
	}

	testcases := []struct {
		add           bool
		lhsCoin       Coin
		lhsMultiAsset map[int]map[string]BigNum
		rhsCoin       Coin
		rhsMultiAsset map[int]map[string]BigNum
		resCoin       Coin
		resMultiAsset map[int]map[string]BigNum
	}{
		{
			add:     true,
			lhsCoin: 10e6,
			lhsMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 10e6, "token2": 20e6},
			},
			rhsCoin: 10e6,
			rhsMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 5e6, "token2": 5e6},
			},
			resCoin: 20e6,
			resMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 15e6, "token2": 25e6},
			},
		},
		{
			add:     true,
			lhsCoin: 10e6,
			lhsMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 10e6},
			},
			rhsCoin: 5e6,
			rhsMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 5e6, "token2": 5e6},
			},
			resCoin: 15e6,
			resMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 15e6, "token2": 5e6},
			},
		},
		{
			add:     true,
			lhsCoin: 10e6,
			lhsMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 10e6},
			},
			rhsCoin: 5e6,
			rhsMultiAsset: map[int]map[string]BigNum{
				0: {"token2": 5e6},
			},
			resCoin: 15e6,
			resMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 10e6, "token2": 5e6},
			},
		},
		{
			add:     true,
			lhsCoin: 10e6,
			lhsMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 10e6},
				1: {"token3": 10e6},
			},
			rhsCoin: 5e6,
			rhsMultiAsset: map[int]map[string]BigNum{
				0: {"token2": 5e6},
				1: {"token3": 10e6},
			},
			resCoin: 15e6,
			resMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 10e6, "token2": 5e6},
				1: {"token3": 20e6},
			},
		},
		{
			add:     false,
			lhsCoin: 10e6,
			lhsMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 10e6, "token2": 20e6},
			},
			rhsCoin: 5e6,
			rhsMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 5e6, "token2": 5e6},
			},
			resCoin: 5e6,
			resMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 5e6, "token2": 15e6},
			},
		},
		{
			add:           false,
			lhsCoin:       10e6,
			lhsMultiAsset: map[int]map[string]BigNum{},
			rhsCoin:       5e6,
			rhsMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 5e6, "token2": 5e6},
			},
			resCoin: 5e6,
			resMultiAsset: map[int]map[string]BigNum{
				0: {"token1": 0, "token2": 0},
			},
		},
		{
			add:           false,
			lhsCoin:       10e6,
			lhsMultiAsset: map[int]map[string]BigNum{},
			rhsCoin:       5e6,
			rhsMultiAsset: map[int]map[string]BigNum{},
			resCoin:       5e6,
			resMultiAsset: map[int]map[string]BigNum{},
		},
	}

	for _, tc := range testcases {
		lhsMultiAssets := NewMultiAsset()
		for policy, assets := range tc.lhsMultiAsset {
			script, err := NewScriptPubKey(pubKeys[policy])
			if err != nil {
				t.Fatal(err)
			}
			policyID, _ := NewPolicyID(script)
			lhsAssets := NewAssets()
			for assetName, value := range assets {
				lhsAssets.Set(NewAssetName(assetName), value)
			}
			lhsMultiAssets.Set(policyID, lhsAssets)
		}

		rhsMultiAssets := NewMultiAsset()
		for policy, assets := range tc.rhsMultiAsset {
			script, err := NewScriptPubKey(pubKeys[policy])
			if err != nil {
				t.Fatal(err)
			}
			policyID, _ := NewPolicyID(script)
			rhsAssets := NewAssets()
			for assetName, value := range assets {
				rhsAssets.Set(NewAssetName(assetName), value)
			}
			rhsMultiAssets.Set(policyID, rhsAssets)
		}

		resMultiAssets := NewMultiAsset()
		for policy, assets := range tc.resMultiAsset {
			script, err := NewScriptPubKey(pubKeys[policy])
			if err != nil {
				t.Fatal(err)
			}
			policyID, _ := NewPolicyID(script)
			resAssets := NewAssets()
			for assetName, value := range assets {
				resAssets.Set(NewAssetName(assetName), value)
			}
			resMultiAssets.Set(policyID, resAssets)
		}

		lhsValue := NewValueWithAssets(tc.lhsCoin, lhsMultiAssets)
		rhsValue := NewValueWithAssets(tc.rhsCoin, rhsMultiAssets)
		resValue := NewValueWithAssets(tc.resCoin, resMultiAssets)

		got := &Value{}
		want := resValue
		if tc.add {
			got = lhsValue.Add(rhsValue)
		} else {
			got = lhsValue.Sub(rhsValue)
		}

		equal, err := got.Cmp(want)
		if err != nil {
			t.Fatal(err)
		}

		if equal != 0 {
			t.Errorf("invalid Add\ngot: %v\nwant: %v", got, want)
		}
	}
}
