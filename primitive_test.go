package cardano

import (
	"errors"
	"testing"
)

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
